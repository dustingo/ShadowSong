# Email 通知渠道 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add email notification channel type supporting SMTP direct send, with global SMTP config and per-route recipients.

**Architecture:** New `EmailSender` implementing the existing `Sender` interface, a global `SmtpConfig` model for SMTP server settings, and `recipients` field on `RouteRule` for per-route email addresses. Follows the same factory/switch pattern as feishu/dingtalk/wecom/webhook.

**Tech Stack:** Go `net/smtp` + `crypto/tls` for SMTP, GORM for SmtpConfig persistence, PrimeReact for frontend forms.

---

## File Structure

| Action | File | Responsibility |
|--------|------|----------------|
| Create | `internal/notifier/email.go` | EmailSender struct, constructor, Send method |
| Modify | `internal/notifier/notifier.go:37-48` | Add `email` case to SendToChannel switch |
| Modify | `internal/models/models.go:92-148` | Add `email` to validTypes, email config validation, SmtpConfig struct, Recipients field on RouteRule |
| Modify | `internal/handlers/config.go:348-372` | Modify TestChannel to accept recipients for email testing |
| Modify | `internal/handlers/config.go:412-435` | Add SmtpConfig CRUD handlers |
| Modify | `internal/handlers/config.go:568-580` | Add `password` handling for email mask (already covered by existing sensitiveKeys) |
| Modify | `internal/router/router.go:140-152` | Add SMTP config routes |
| Modify | `internal/database/postgres.go:47-58` | Add SmtpConfig to AutoMigrate tables |
| Modify | `frontend/src/types/index.ts:89-107` | Add `email` to Channel type union, SmtpConfig type, recipients to RouteRule |
| Modify | `frontend/src/pages/Channels.tsx:21-24,190-195` | Add email to CHANNEL_TYPES / channelTypeOptions, email config form fields |
| Modify | `frontend/src/pages/Channels.tsx:274-285` | Modify handleTest for email to prompt recipients |
| Modify | `frontend/src/pages/RouteRules.tsx:21-31,387-397` | Add recipients to RouteRuleFormData, show recipients input when email channel selected |
| Modify | `frontend/src/api/client.ts:138-155` | Add smtpConfigApi, update channelApi.test, add routeRuleApi with recipients |
| Modify | `frontend/src/stores/configStore.ts` | Add SMTP config state and actions |
| Create | `frontend/src/pages/SmtpSettings.tsx` | SMTP configuration page |
| Modify | `frontend/src/App.tsx` | Add SMTP settings route |
| Modify | `frontend/src/components/layout/AppSidebar.tsx` | Add SMTP settings nav item |

---

### Task 1: SmtpConfig Model and AutoMigrate

**Files:**
- Modify: `internal/models/models.go:150-166` (after RouteRule struct)
- Modify: `internal/database/postgres.go:47-58`

- [ ] **Step 1: Add SmtpConfig struct to models.go**

Add after the `RouteRule` struct (after line 166):

```go
// SmtpConfig represents global SMTP server configuration
type SmtpConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Host      string    `gorm:"size:128;not null" json:"host"`
	Port      int       `gorm:"not null;default:465" json:"port"`
	Username  string    `gorm:"size:128;not null" json:"username"`
	Password  string    `gorm:"size:256" json:"password"`
	FromAddr  string    `gorm:"size:128;not null" json:"from_addr"`
	FromName  string    `gorm:"size:64" json:"from_name"`
	TLS       bool      `gorm:"default:true" json:"tls"`
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *SmtpConfig) Validate() error {
	if s.Host == "" {
		return errors.New("host is required")
	}
	if s.Port == 0 {
		return errors.New("port is required")
	}
	if s.Username == "" {
		return errors.New("username is required")
	}
	if s.FromAddr == "" {
		return errors.New("from_addr is required")
	}
	return nil
}
```

- [ ] **Step 2: Add Recipients field to RouteRule struct**

Add after the `ChannelIDs` field (line 158):

```go
Recipients datatypes.JSON `json:"recipients"` // []string — email addresses for email channels
```

Also add default in `BeforeCreate` (after line 193):

```go
if r.Recipients == nil {
	r.Recipients = datatypes.JSON("[]")
}
```

- [ ] **Step 3: Add `email` to Channel validTypes**

Modify the `validTypes` map in `Channel.Validate()` (lines 99-106):

```go
validTypes := map[string]bool{
	"feishu":   true,
	"dingtalk": true,
	"wecom":    true,
	"webhook":  true,
	"email":    true,
}
```

Add email config validation after the webhook block (after line 146):

```go
if c.Type == "email" {
	var cfg struct {
		FromName string `json:"from_name"`
	}
	json.Unmarshal(c.Config, &cfg)
	// email channel config is minimal; from_name is optional
	// SMTP connection and recipients are validated at send time
}
```

- [ ] **Step 4: Add SmtpConfig to AutoMigrate**

In `internal/database/postgres.go`, add `&models.SmtpConfig{}` to the tables slice (after line 57):

```go
&models.NotificationDeliveryRecovery{},
&models.SmtpConfig{},
```

- [ ] **Step 5: Run the application to verify migration**

Run: `cd d:/goproject/shadowsongAI && go build ./...`
Expected: Compiles successfully with no errors.

- [ ] **Step 6: Commit**

```bash
git add internal/models/models.go internal/database/postgres.go
git commit -m "feat(email): add SmtpConfig model and email channel type"
```

---

### Task 2: EmailSender Implementation

**Files:**
- Create: `internal/notifier/email.go`
- Modify: `internal/notifier/notifier.go:37-48`

- [ ] **Step 1: Create EmailSender in email.go**

Create `internal/notifier/email.go`:

```go
package notifier

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"gorm.io/gorm"
)

var globalDB *gorm.DB

// SetDB sets the global DB reference used by EmailSender to load SmtpConfig.
// Must be called once at application startup after DB initialization.
func SetDB(db *gorm.DB) {
	globalDB = db
}

type EmailConfig struct {
	FromName string `json:"from_name"`
}

type EmailSender struct {
	config EmailConfig
	db     *gorm.DB
}

func NewEmailSender(config []byte) (Sender, error) {
	var ec EmailConfig
	if err := json.Unmarshal(config, &ec); err != nil {
		return nil, err
	}
	return &EmailSender{config: ec, db: globalDB}, nil
}

func (s *EmailSender) Send(title, content string, data map[string]interface{}) error {
	if s.db == nil {
		return fmt.Errorf("SMTP 服务未配置: 数据库未初始化")
	}

	var smtpCfg models.SmtpConfig
	if err := s.db.Where("id = 1").First(&smtpCfg).Error; err != nil {
		return fmt.Errorf("SMTP 服务未配置: %w", err)
	}
	if !smtpCfg.Enabled {
		return fmt.Errorf("SMTP 服务未启用")
	}

	recipientsRaw, ok := data["recipients"]
	if !ok || recipientsRaw == nil {
		return fmt.Errorf("邮件收件人为空")
	}

	var recipients []string
	switch v := recipientsRaw.(type) {
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				recipients = append(recipients, s)
			}
		}
	case []string:
		recipients = v
	case string:
		if v != "" {
			recipients = []string{v}
		}
	}
	if len(recipients) == 0 {
		return fmt.Errorf("邮件收件人为空")
	}

	fromName := s.config.FromName
	if fromName == "" {
		fromName = smtpCfg.FromName
	}
	if fromName == "" {
		fromName = "告警系统"
	}

	from := smtpCfg.FromAddr
	subject := encodeRFC2047(title)
	message := buildHTMLMessage(from, fromName, recipients, subject, content)

	addr := fmt.Sprintf("%s:%d", smtpCfg.Host, smtpCfg.Port)
	auth := smtp.PlainAuth("", smtpCfg.Username, smtpCfg.Password, smtpCfg.Host)

	var client *smtp.Client
	var err error

	if smtpCfg.TLS {
		tlsConfig := &tls.Config{
			ServerName: smtpCfg.Host,
		}
		conn, dialErr := tls.DialWithDialer(
			&net.Dialer{Timeout: 10 * time.Second},
			"tcp",
			addr,
			tlsConfig,
		)
		if dialErr != nil {
			return fmt.Errorf("failed to connect SMTP server: %w", dialErr)
		}
		client, err = smtp.NewClient(conn, smtpCfg.Host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
	} else {
		client, err = smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("failed to dial SMTP server: %w", err)
		}
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL FROM failed: %w", err)
	}
	for _, rcpt := range recipients {
		if err := client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("SMTP RCPT TO failed for %s: %w", rcpt, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA failed: %w", err)
	}
	if _, err := w.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close email body: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("SMTP QUIT failed: %w", err)
	}

	return nil
}

func encodeRFC2047(s string) string {
	for _, r := range s {
		if r > 127 {
			encoded := base64.StdEncoding.EncodeToString([]byte(s))
			return fmt.Sprintf("=?UTF-8?B?%s?=", encoded)
		}
	}
	return s
}

func buildHTMLMessage(from, fromName string, to []string, subject, htmlBody string) string {
	var buf strings.Builder

	fromAddr := mail.Address{Name: fromName, Address: from}
	buf.WriteString(fmt.Sprintf("From: %s\r\n", fromAddr.String()))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(htmlBody)

	return buf.String()
}
```

- [ ] **Step 2: Add email case to SendToChannel switch**

In `internal/notifier/notifier.go`, add after the webhook case (line 45):

```go
case "email":
	sender, err = NewEmailSender(configBytes)
```

- [ ] **Step 3: Add notifier.SetDB call in application startup**

Find the main application entry point and add `notifier.SetDB(db)` after db initialization. Search for where `database.InitDB` is called, and add `notifier.SetDB(db)` right after.

- [ ] **Step 4: Build to verify compilation**

Run: `cd d:/goproject/shadowsongAI && go build ./...`
Expected: Compiles successfully.

- [ ] **Step 5: Commit**

```bash
git add internal/notifier/email.go internal/notifier/notifier.go
git commit -m "feat(email): add EmailSender implementation"
```

---

### Task 3: SMTP Config API Handlers

**Files:**
- Modify: `internal/handlers/config.go:348-372,412-435,568-580`

- [ ] **Step 1: Add SmtpConfig CRUD handlers to config.go**

Add after the TestChannel function (after line 372):

```go
// ============ SmtpConfig ============

func (h *ConfigHandler) GetSmtpConfig(c *gin.Context) {
	var cfg models.SmtpConfig
	err := h.db.Where("id = 1").First(&cfg).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	cfg.Password = "****"
	c.JSON(http.StatusOK, cfg)
}

func (h *ConfigHandler) UpdateSmtpConfig(c *gin.Context) {
	var cfg models.SmtpConfig
	err := h.db.Where("id = 1").First(&cfg).Error

	isNew := err == gorm.ErrRecordNotFound

	var input models.SmtpConfig
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := input.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if isNew {
		input.ID = 1
		if err := h.db.Create(&input).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		_ = recordAudit(h.db, c, "config.smtp.update", "smtp_config", "1", auditResultAllowed, "created SMTP config")
		input.Password = "****"
		c.JSON(http.StatusOK, input)
		return
	}

	cfg.Host = input.Host
	cfg.Port = input.Port
	cfg.Username = input.Username
	cfg.FromAddr = input.FromAddr
	cfg.FromName = input.FromName
	cfg.TLS = input.TLS
	cfg.Enabled = input.Enabled

	// Only update password if not masked
	if input.Password != "" && input.Password != "****" {
		cfg.Password = input.Password
	}

	if err := h.db.Save(&cfg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = recordAudit(h.db, c, "config.smtp.update", "smtp_config", "1", auditResultAllowed, "updated SMTP config")
	cfg.Password = "****"
	c.JSON(http.StatusOK, cfg)
}

func (h *ConfigHandler) TestSmtpConfig(c *gin.Context) {
	var input struct {
		Recipients []string `json:"recipients" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recipients is required"})
		return
	}

	ch := &models.Channel{
		Type:   "email",
		Config: datatypes.JSON(`{"from_name":"告警系统"}`),
	}

	data := map[string]interface{}{
		"recipients": input.Recipients,
	}

	if err := notifier.SendToChannel(ch, "测试通知", "这是一条来自游戏运维告警系统的测试消息。", data); err != nil {
		_ = recordAudit(h.db, c, "config.smtp.test", "smtp_config", "1", auditResultDenied, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = recordAudit(h.db, c, "config.smtp.test", "smtp_config", "1", auditResultAllowed, "SMTP test email sent")
	c.JSON(http.StatusOK, gin.H{"message": "test email sent successfully"})
}
```

Note: `notifier.SendToChannel` will work because `notifier.SetDB(db)` is called at app startup, making `globalDB` available to `EmailSender`.

- [ ] **Step 2: Modify TestChannel to accept recipients for email**

In `internal/handlers/config.go`, modify the `TestChannel` function (lines 348-372) to parse an optional request body:

Replace the function body:

```go
func (h *ConfigHandler) TestChannel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var ch models.Channel
	if err := h.db.First(&ch, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	if !ch.Enabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "channel is disabled"})
		return
	}

	// Parse optional request body for email recipients
	var body struct {
		Recipients []string `json:"recipients"`
	}
	c.ShouldBindJSON(&body)

	testTitle := "测试通知"
	testContent := "这是一条来自游戏运维告警系统的测试消息。"

	data := map[string]interface{}{}
	if len(body.Recipients) > 0 {
		data["recipients"] = body.Recipients
	}

	if err := notifier.SendToChannel(&ch, testTitle, testContent, data); err != nil {
		_ = recordAudit(h.db, c, "config.channel.test", "channel", strconv.FormatUint(uint64(ch.ID), 10), auditResultDenied, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	_ = recordAudit(h.db, c, "config.channel.test", "channel", strconv.FormatUint(uint64(ch.ID), 10), auditResultAllowed, "test notification sent")
	c.JSON(http.StatusOK, gin.H{"message": "test notification sent successfully"})
}
```

- [ ] **Step 3: Add recipients to RouteRule Update handler**

In `internal/handlers/config.go`, modify the `UpdateRouteRule` function (lines 412-435) to include `Recipients`:

After line 428 (`rule.ChannelIDs = input.ChannelIDs`), add:

```go
rule.Recipients = input.Recipients
```

And in `CreateRouteRule` (after line 396), ensure `Recipients` is included in the bound input — it already is since we bind `models.RouteRule` which now has the field.

- [ ] **Step 4: Build to verify**

Run: `cd d:/goproject/shadowsongAI && go build ./...`
Expected: Compiles successfully.

- [ ] **Step 5: Commit**

```bash
git add internal/handlers/config.go
git commit -m "feat(email): add SMTP config handlers and email test support"
```

---

### Task 4: SMTP Config API Routes

**Files:**
- Modify: `internal/router/router.go:140-165`

- [ ] **Step 1: Add SMTP config routes**

In `internal/router/router.go`, add after the channel routes block (after line 152):

```go
// SMTP config routes (protected)
smtp := v1.Group("/smtp-config")
smtp.Use(middleware.JWTAuth(jwtAuth, db))
{
	smtp.GET("", middleware.RequireCapability(authz.CapabilityViewConfig), configHandler.GetSmtpConfig)
	smtp.PUT("", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.UpdateSmtpConfig)
	smtp.POST("/test", middleware.RequireCapability(authz.CapabilityManageConfig), configHandler.TestSmtpConfig)
}
```

- [ ] **Step 2: Build to verify**

Run: `cd d:/goproject/shadowsongAI && go build ./...`
Expected: Compiles successfully.

- [ ] **Step 3: Commit**

```bash
git add internal/router/router.go
git commit -m "feat(email): add SMTP config API routes"
```

---

### Task 5: Inject Recipients into Notification Data Flow

**Files:**
- Modify: `internal/handlers/webhook.go:952-971`

- [ ] **Step 1: Pass recipients from RouteRule to notification data**

In `internal/handlers/webhook.go`, modify the `processAlertNotificationsWithHook` function. After line 963 (the `for _, target := range matchedTargets` loop), modify the `sendNotification` call area.

Find the `buildNotificationRenderContext` call inside `sendChannelNotification` (around line 1160) and add recipients:

In the `sendChannelNotification` function, after line 1160 (`data := h.buildNotificationRenderContext(alert)`), add:

```go
// Inject recipients from route rule for email channels
if routeRule != nil && len(routeRule.Recipients) > 0 {
	var recipients []string
	if err := json.Unmarshal(routeRule.Recipients, &recipients); err == nil && len(recipients) > 0 {
		data["recipients"] = recipients
	}
}
```

- [ ] **Step 2: Build to verify**

Run: `cd d:/goproject/shadowsongAI && go build ./...`
Expected: Compiles successfully.

- [ ] **Step 3: Commit**

```bash
git add internal/handlers/webhook.go
git commit -m "feat(email): inject recipients from route rule into notification data"
```

---

### Task 6: Frontend Type Updates

**Files:**
- Modify: `frontend/src/types/index.ts:89-107,109-124`

- [ ] **Step 1: Add email to Channel type union and SmtpConfig type**

In `frontend/src/types/index.ts`:

Update the Channel type (line 92):

```typescript
type: 'feishu' | 'dingtalk' | 'wecom' | 'webhook' | 'email'
```

Add email config fields to the Channel config interface (after line 102):

```typescript
config: JsonObject & {
  webhook_url?: string
  secret?: string
  url?: string
  method?: string
  content_type?: string
  headers?: Record<string, string> | string
  template?: string
  auth_type?: string
  auth_config?: WebhookAuthConfig
  from_name?: string  // email channel
}
```

Add SmtpConfig interface after the Channel interface (after line 107):

```typescript
export interface SmtpConfig {
  id?: number
  host: string
  port: number
  username: string
  password: string
  from_addr: string
  from_name: string
  tls: boolean
  enabled: boolean
  updated_at?: string
}
```

Update RouteRule interface (line 116) to add recipients:

```typescript
recipients: string[]
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/types/index.ts
git commit -m "feat(email): add SmtpConfig type and email to Channel type"
```

---

### Task 7: Frontend API and Store Updates

**Files:**
- Modify: `frontend/src/api/client.ts:136-155,157-173`
- Modify: `frontend/src/stores/configStore.ts`

- [ ] **Step 1: Add smtpConfigApi and update channelApi.test**

In `frontend/src/api/client.ts`:

Update channelApi test method (line 154):

```typescript
test: (id: number, body?: { recipients?: string[] }) =>
  apiClient.post(`/channels/${id}/test`, body),
```

Add smtpConfigApi after channelApi (after line 155):

```typescript
// ============ SmtpConfig API ============

export const smtpConfigApi = {
  get: () => apiClient.get<SmtpConfig>('/smtp-config'),

  update: (data: Partial<SmtpConfig>) =>
    apiClient.put<SmtpConfig>('/smtp-config', data),

  test: (recipients: string[]) =>
    apiClient.post('/smtp-config/test', { recipients }),
}
```

Make sure to import `SmtpConfig` from types.

- [ ] **Step 2: Add SMTP config state and actions to configStore**

In `frontend/src/stores/configStore.ts`:

Add SmtpConfig to the type imports.

Add to the ConfigState interface:

```typescript
smtpConfig: SmtpConfig | null
smtpConfigLoading: boolean
fetchSmtpConfig: () => Promise<void>
updateSmtpConfig: (data: Partial<SmtpConfig>) => Promise<void>
testSmtpConfig: (recipients: string[]) => Promise<void>
```

Add to the store implementation:

```typescript
smtpConfig: null,
smtpConfigLoading: false,

fetchSmtpConfig: async () => {
  set({ smtpConfigLoading: true })
  try {
    const data = await smtpConfigApi.get() as unknown as SmtpConfig
    set({ smtpConfig: data, smtpConfigLoading: false })
  } catch (error) {
    set({ smtpConfigLoading: false })
    throw error
  }
},

updateSmtpConfig: async (data) => {
  await smtpConfigApi.update(data)
  get().fetchSmtpConfig()
},

testSmtpConfig: async (recipients) => {
  await smtpConfigApi.test(recipients)
},
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/api/client.ts frontend/src/stores/configStore.ts
git commit -m "feat(email): add SMTP config API client and store"
```

---

### Task 8: Frontend SMTP Settings Page

**Files:**
- Create: `frontend/src/pages/SmtpSettings.tsx`
- Modify: `frontend/src/App.tsx:6-17,87-101`
- Modify: `frontend/src/components/layout/AppSidebar.tsx`

- [ ] **Step 1: Create SmtpSettings page**

Create `frontend/src/pages/SmtpSettings.tsx`:

```tsx
import React, { useEffect, useState } from 'react'
import { Card } from 'primereact/card'
import { Button } from 'primereact/button'
import { InputText } from 'primereact/inputtext'
import { InputNumber } from 'primereact/inputnumber'
import { InputSwitch } from 'primereact/inputswitch'
import { Dialog } from 'primereact/dialog'
import { useToast, PermissionNotice } from '../components'
import { canUser, capabilityManageConfig, isReadOnlyConfigUser } from '../authz/capabilities'
import { useConfigStore } from '../stores/configStore'
import { useUserStore } from '../stores/userStore'
import { getApiErrorMessage } from '../api/client'
import type { SmtpConfig } from '../types'

export const SmtpSettings: React.FC = () => {
  const user = useUserStore((state) => state.user)
  const toast = useToast()
  const { smtpConfig, smtpConfigLoading, fetchSmtpConfig, updateSmtpConfig, testSmtpConfig } = useConfigStore()

  const canManageConfig = canUser(user, capabilityManageConfig)
  const readOnly = isReadOnlyConfigUser(user)

  const [formValues, setFormValues] = useState<Partial<SmtpConfig>>({
    host: '',
    port: 465,
    username: '',
    password: '',
    from_addr: '',
    from_name: '',
    tls: true,
    enabled: true,
  })
  const [testDialogVisible, setTestDialogVisible] = useState(false)
  const [testRecipient, setTestRecipient] = useState('')

  useEffect(() => {
    fetchSmtpConfig()
  }, [fetchSmtpConfig])

  useEffect(() => {
    if (smtpConfig) {
      setFormValues({
        host: smtpConfig.host || '',
        port: smtpConfig.port || 465,
        username: smtpConfig.username || '',
        password: '****',
        from_addr: smtpConfig.from_addr || '',
        from_name: smtpConfig.from_name || '',
        tls: smtpConfig.tls ?? true,
        enabled: smtpConfig.enabled ?? true,
      })
    }
  }, [smtpConfig])

  const handleSave = async () => {
    if (!canManageConfig) {
      toast.showWarn('当前角色无权执行该操作')
      return
    }
    try {
      await updateSmtpConfig(formValues)
      toast.showSuccess('保存成功')
    } catch (error) {
      toast.showError(getApiErrorMessage(error, '保存失败'))
    }
  }

  const handleTest = async () => {
    if (!testRecipient.trim()) {
      toast.showError('请输入收件人地址')
      return
    }
    try {
      await testSmtpConfig([testRecipient.trim()])
      toast.showSuccess('测试邮件已发送')
      setTestDialogVisible(false)
      setTestRecipient('')
    } catch (error) {
      toast.showError(getApiErrorMessage(error, '发送失败'))
    }
  }

  const cardHeader = (
    <div className="flex align-items-center justify-content-between">
      <div>
        <span className="text-xl font-bold">邮件服务配置</span>
        <span className="text-color-secondary text-sm ml-2">配置 SMTP 服务器用于发送告警邮件</span>
      </div>
      <div className="flex gap-2">
        <Button label="测试连接" icon="pi pi-send" outlined onClick={() => setTestDialogVisible(true)} disabled={!canManageConfig || !smtpConfig?.host} />
        <Button label="保存" icon="pi pi-check" onClick={handleSave} disabled={!canManageConfig} loading={smtpConfigLoading} />
      </div>
    </div>
  )

  return (
    <div>
      <Card className="shadow-sm border-0" header={cardHeader}>
        {readOnly && (
          <PermissionNotice
            title="当前角色可查看配置，但不能修改"
            description="邮件服务配置对非 `admin` 角色保持只读。"
            type="info"
          />
        )}
        <div className="flex flex-column gap-3 p-fluid" style={{ maxWidth: '600px' }}>
          <div className="field">
            <label className="font-medium">SMTP 服务器地址</label>
            <InputText
              value={formValues.host || ''}
              onChange={(e) => setFormValues({ ...formValues, host: e.target.value })}
              placeholder="smtp.example.com"
              disabled={!canManageConfig}
            />
          </div>
          <div className="field">
            <label className="font-medium">端口</label>
            <InputNumber
              value={formValues.port}
              onValueChange={(e) => setFormValues({ ...formValues, port: e.value ?? 465 })}
              min={1}
              max={65535}
              disabled={!canManageConfig}
            />
          </div>
          <div className="field">
            <label className="font-medium">用户名</label>
            <InputText
              value={formValues.username || ''}
              onChange={(e) => setFormValues({ ...formValues, username: e.target.value })}
              placeholder="user@example.com"
              disabled={!canManageConfig}
            />
          </div>
          <div className="field">
            <label className="font-medium">密码</label>
            <InputText
              type="password"
              value={formValues.password || ''}
              onChange={(e) => setFormValues({ ...formValues, password: e.target.value })}
              placeholder="密码"
              disabled={!canManageConfig}
            />
          </div>
          <div className="field">
            <label className="font-medium">发件人地址</label>
            <InputText
              value={formValues.from_addr || ''}
              onChange={(e) => setFormValues({ ...formValues, from_addr: e.target.value })}
              placeholder="alert@example.com"
              disabled={!canManageConfig}
            />
          </div>
          <div className="field">
            <label className="font-medium">发件人显示名（可选）</label>
            <InputText
              value={formValues.from_name || ''}
              onChange={(e) => setFormValues({ ...formValues, from_name: e.target.value })}
              placeholder="告警系统"
              disabled={!canManageConfig}
            />
          </div>
          <div className="field flex align-items-center gap-2">
            <label className="font-medium mb-0">启用 TLS</label>
            <InputSwitch
              checked={formValues.tls ?? true}
              onChange={(e) => setFormValues({ ...formValues, tls: e.value })}
              disabled={!canManageConfig}
            />
          </div>
          <div className="field flex align-items-center gap-2">
            <label className="font-medium mb-0">启用</label>
            <InputSwitch
              checked={formValues.enabled ?? true}
              onChange={(e) => setFormValues({ ...formValues, enabled: e.value })}
              disabled={!canManageConfig}
            />
          </div>
        </div>
      </Card>

      <Dialog
        header="测试邮件发送"
        visible={testDialogVisible}
        onHide={() => setTestDialogVisible(false)}
        style={{ width: '400px' }}
        footer={
          <div className="flex justify-content-end gap-2">
            <Button label="取消" outlined onClick={() => setTestDialogVisible(false)} />
            <Button label="发送" onClick={handleTest} />
          </div>
        }
      >
        <div className="flex flex-column gap-2 p-fluid">
          <label className="font-medium">收件人地址</label>
          <InputText
            value={testRecipient}
            onChange={(e) => setTestRecipient(e.target.value)}
            placeholder="test@example.com"
          />
        </div>
      </Dialog>
    </div>
  )
}
```

- [ ] **Step 2: Add SmtpSettings to pages index export**

Find and update `frontend/src/pages/index.ts` to add `SmtpSettings` export.

- [ ] **Step 3: Add SMTP settings route to App.tsx**

In `frontend/src/App.tsx`:

Add `SmtpSettings` to the imports (line 6-17).

Add route after the channels route (line 94):

```tsx
<Route path="/smtp-settings" element={<RequireAuth requiredCapability={capabilityViewConfig}><SmtpSettings /></RequireAuth>} />
```

- [ ] **Step 4: Add nav item to AppSidebar**

In `frontend/src/components/layout/AppSidebar.tsx`, find the menuItems array and add after the Channels entry:

```tsx
{ label: '邮件服务', icon: Mail, path: '/smtp-settings' },
```

Make sure to import the `Mail` icon from lucide-react (or use an appropriate icon from the existing icon set).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/SmtpSettings.tsx frontend/src/pages/index.ts frontend/src/App.tsx frontend/src/components/layout/AppSidebar.tsx
git commit -m "feat(email): add SMTP settings page and navigation"
```

---

### Task 9: Frontend Email Channel Form

**Files:**
- Modify: `frontend/src/pages/Channels.tsx:21-24,190-195,51-108,110-164,274-285,402-675`

- [ ] **Step 1: Add email to channelTypeOptions**

In `frontend/src/pages/Channels.tsx`, add to the `channelTypeOptions` array (after line 194):

```tsx
{ value: 'email', label: '邮件', icon: 'pi pi-envelope' },
```

- [ ] **Step 2: Add email config to ChannelFormValues interface**

Add `from_name` to the `ChannelFormValues` config interface (after line 38):

```typescript
from_name?: string  // email channel
```

- [ ] **Step 3: Add email case to formatChannelConfigForForm**

Add after the `wecom` block (after line 80):

```tsx
if (channel.type === 'email') {
  return {
    ...channel,
    config: {
      from_name: String(config.from_name ?? ''),
    },
  }
}
```

- [ ] **Step 4: Add email case to buildChannelPayload**

Add after the `wecom` block (after line 160):

```tsx
if (values.type === 'email') {
  return {
    ...values,
    config: {
      from_name: String(config.from_name ?? ''),
    },
  }
}
```

- [ ] **Step 5: Add email config form fields in renderConfigFields**

Add after the `wecom` block (after line 492), before the `webhook` block:

```tsx
if (type === 'email') {
  return (
    <div className="flex flex-column gap-2">
      <label className="text-sm">发件人显示名（可选，覆盖全局配置）</label>
      <InputText
        placeholder="告警系统"
        value={formValues.config.from_name || ''}
        onChange={(e) =>
          setFormValues({
            ...formValues,
            config: { ...formValues.config, from_name: e.target.value },
          })
        }
        disabled={!canManageConfig}
      />
    </div>
  )
}
```

- [ ] **Step 6: Modify handleTest for email channels**

Modify the `handleTest` function (around line 274) to handle email testing:

```tsx
const handleTest = async (record: Channel) => {
  if (!canManageConfig) {
    toast.showWarn('当前角色无权执行该操作')
    return
  }
  if (record.type === 'email') {
    setTestTargetChannelId(record.id)
    setTestRecipientInput('')
    setTestDialogVisible(true)
    return
  }
  try {
    await testChannel(record.id)
    toast.showSuccess('测试消息已发送')
  } catch (error) {
    toast.showError(getApiErrorMessage(error, '发送失败'))
  }
}
```

Add state variables for the test dialog:

```tsx
const [testDialogVisible, setTestDialogVisible] = useState(false)
const [testTargetChannelId, setTestTargetChannelId] = useState<number | null>(null)
const [testRecipientInput, setTestRecipientInput] = useState('')
```

Add the test dialog JSX and handler:

```tsx
const handleTestEmail = async () => {
  if (!testTargetChannelId) return
  if (!testRecipientInput.trim()) {
    toast.showError('请输入收件人地址')
    return
  }
  try {
    await testChannel(testTargetChannelId, { recipients: [testRecipientInput.trim()] })
    toast.showSuccess('测试邮件已发送')
    setTestDialogVisible(false)
  } catch (error) {
    toast.showError(getApiErrorMessage(error, '发送失败'))
  }
}
```

Add the Dialog component at the end of the JSX:

```tsx
<Dialog
  header="测试邮件发送"
  visible={testDialogVisible}
  onHide={() => setTestDialogVisible(false)}
  style={{ width: '400px' }}
  footer={
    <div className="flex justify-content-end gap-2">
      <Button label="取消" outlined onClick={() => setTestDialogVisible(false)} />
      <Button label="发送" onClick={handleTestEmail} />
    </div>
  }
>
  <div className="flex flex-column gap-2 p-fluid">
    <label className="font-medium">收件人地址</label>
    <InputText
      value={testRecipientInput}
      onChange={(e) => setTestRecipientInput(e.target.value)}
      placeholder="test@example.com"
    />
  </div>
</Dialog>
```

- [ ] **Step 7: Commit**

```bash
git add frontend/src/pages/Channels.tsx
git commit -m "feat(email): add email channel type to channels page"
```

---

### Task 10: Frontend RouteRule Recipients Field

**Files:**
- Modify: `frontend/src/pages/RouteRules.tsx:21-31,387-397`

- [ ] **Step 1: Add recipients to RouteRuleFormData**

In `frontend/src/pages/RouteRules.tsx`, add to the `RouteRuleFormData` interface (after line 26):

```typescript
recipients: string[]
```

Set default in `resetForm` (around line 97):

```typescript
recipients: [],
```

Set value in `handleEdit` (around line 128):

```typescript
recipients: record.recipients || [],
```

- [ ] **Step 2: Add recipients input to route rule form dialog**

After the `channel_ids` MultiSelect field (around line 397), add:

```tsx
{formData.channel_ids?.some((cid) => {
  const ch = channels.find((c) => c.id === cid)
  return ch?.type === 'email'
}) && (
  <div className="field">
    <label htmlFor="recipients" className="font-medium">邮件收件人</label>
    <div className="flex flex-wrap gap-2 mb-2">
      {formData.recipients?.map((r, idx) => (
        <Tag
          key={idx}
          value={r}
          icon="pi pi-envelope"
          severity="info"
          className="cursor-pointer"
          onClick={() => {
            setFormData({
              ...formData,
              recipients: formData.recipients.filter((_, i) => i !== idx),
            })
          }}
          style={{ cursor: 'pointer' }}
        />
      ))}
    </div>
    <div className="flex gap-2">
      <InputText
        id="recipients"
        placeholder="输入邮箱地址后回车添加"
        onKeyDown={(e) => {
          if (e.key === 'Enter') {
            e.preventDefault()
            const val = (e.target as HTMLInputElement).value.trim()
            if (val && val.includes('@') && !formData.recipients?.includes(val)) {
              setFormData({
                ...formData,
                recipients: [...(formData.recipients || []), val],
              })
              ;(e.target as HTMLInputElement).value = ''
            }
          }
        }}
        disabled={!canManageConfig}
      />
    </div>
    <small style={{ color: 'var(--text-secondary)' }}>
      当关联渠道包含邮件类型时，需填写收件人地址
    </small>
  </div>
)}
```

Make sure to import `Tag` from `primereact/tag` (already imported at line 8).

- [ ] **Step 3: Add validation for recipients when email channel is selected**

In the `handleSubmit` function (around line 167), add after the channel_ids check:

```tsx
const hasEmailChannel = formData.channel_ids?.some((cid) => {
  const ch = channels.find((c) => c.id === cid)
  return ch?.type === 'email'
})
if (hasEmailChannel && (!formData.recipients || formData.recipients.length === 0)) {
  toast.showError('关联邮件渠道时必须填写收件人地址')
  return
}
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/RouteRules.tsx
git commit -m "feat(email): add recipients field to route rules form"
```

---

### Task 11: Integration Testing

**Files:**
- No new files — manual verification

- [ ] **Step 1: Start the backend**

Run: `cd d:/goproject/shadowsongAI && go run ./cmd/server` (or the appropriate start command)
Expected: Server starts without errors, SmtpConfig table created.

- [ ] **Step 2: Start the frontend**

Run: `cd d:/goproject/shadowsongAI/frontend && pnpm dev`
Expected: Frontend dev server starts.

- [ ] **Step 3: Verify SMTP settings page**

1. Navigate to "邮件服务" in sidebar
2. Fill in SMTP config (host, port, username, password, from address)
3. Click "测试连接" — enter a recipient email, verify test email is sent
4. Click "保存" — verify config is saved and password is masked on reload

- [ ] **Step 4: Verify email channel creation**

1. Navigate to "推送渠道管理"
2. Click "新建渠道", select "邮件" type
3. Verify only "发件人显示名" field appears
4. Create the email channel

- [ ] **Step 5: Verify route rule recipients**

1. Navigate to "路由规则管理"
2. Create/edit a rule that includes an email channel
3. Verify "邮件收件人" input appears
4. Add email addresses, save

- [ ] **Step 6: Verify end-to-end alert flow**

1. Send a test webhook that matches a route rule with email channel
2. Verify email is delivered to the configured recipients
3. Check delivery record in "投递历史" page

- [ ] **Step 7: Commit any fixups**

```bash
git add -A
git commit -m "fix(email): integration test fixups"
```

---

### Task 12: Final Commit

- [ ] **Step 1: Review all changes**

Run: `cd d:/goproject/shadowsongAI && git diff --stat main`
Expected: Only files from this feature changed.

- [ ] **Step 2: Run full test suite**

Run: `cd d:/goproject/shadowsongAI && go test ./...`
Expected: All tests pass.

Run: `cd d:/goproject/shadowsongAI/frontend && pnpm test`
Expected: All frontend tests pass.