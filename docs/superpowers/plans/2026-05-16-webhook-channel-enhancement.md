# Webhook Channel Enhancement Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enhance the webhook channel type to support `application/x-www-form-urlencoded` body format, HTTP Basic Auth, and custom header authentication — enabling integration with external alerting platforms like Odin.

**Architecture:** Extend the existing `WebhookConfig` struct in the notifier package with new fields (`content_type`, `auth_type`, `auth_config`). The sender logic branches on these fields to construct the appropriate HTTP request. Frontend adds conditional form fields for the new config options. All changes are backward-compatible — missing fields default to current behavior.

**Tech Stack:** Go 1.25 + Gin + GORM (backend), React 18 + TypeScript + PrimeReact (frontend)

---

## File Structure

| File | Responsibility |
|------|---------------|
| `internal/notifier/webhook.go` | Webhook sender: parse enhanced config, build request with content_type/auth support |
| `internal/notifier/notifier_test.go` | Unit tests for WebhookSender with new config fields |
| `internal/models/models.go` | Channel.Validate() — add webhook config field validation |
| `internal/handlers/config.go` | maskChannelConfig — mask `password` and `header_value` in auth_config |
| `frontend/src/types/index.ts` | Extend Channel type with webhook config fields |
| `frontend/src/pages/Channels.tsx` | Webhook config form: add content_type, auth_type, auth_config fields |

---

### Task 1: Backend — Extend WebhookConfig and Sender Logic

**Files:**
- Modify: `internal/notifier/webhook.go` (lines 306-370)

- [ ] **Step 1: Add new config structs**

Replace the existing `WebhookConfig` struct (lines 306-311) with the enhanced version. Add `BasicAuthConfig` and `CustomAuthConfig` structs after it.

```go
type WebhookConfig struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	ContentType string            `json:"content_type"`
	Headers     map[string]string `json:"headers"`
	Template    string            `json:"template"`
	AuthType    string            `json:"auth_type"`
	AuthConfig  json.RawMessage   `json:"auth_config"`
}

type BasicAuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CustomAuthConfig struct {
	HeaderName  string `json:"header_name"`
	HeaderValue string `json:"header_value"`
}
```

- [ ] **Step 2: Update NewWebhookSender to set defaults for new fields**

Replace the existing `NewWebhookSender` function (lines 318-333). Add defaults for `ContentType` and `AuthType`.

```go
func NewWebhookSender(config json.RawMessage) (Sender, error) {
	var wc WebhookConfig
	if err := json.Unmarshal(config, &wc); err != nil {
		return nil, err
	}
	if wc.URL == "" {
		return nil, fmt.Errorf("webhook url is required")
	}
	if wc.Method == "" {
		wc.Method = "POST"
	}
	if wc.ContentType == "" {
		wc.ContentType = "application/json"
	}
	if wc.AuthType == "" {
		wc.AuthType = "none"
	}
	return &WebhookSender{
		config: wc,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}
```

- [ ] **Step 3: Rewrite WebhookSender.Send to support content_type and auth**

Replace the existing `Send` method (lines 335-370). The new version handles content_type branching and auth.

```go
func (s *WebhookSender) Send(title, content string) error {
	var bodyReader io.Reader

	if s.config.ContentType == "application/x-www-form-urlencoded" {
		// For form-urlencoded, send content string directly as body
		bodyReader = strings.NewReader(content)
	} else {
		// JSON mode (default)
		var body []byte
		if s.config.Template != "" {
			body = []byte(s.config.Template)
		} else {
			payload := map[string]string{
				"title":   title,
				"content": content,
			}
			body, _ = json.Marshal(payload)
		}
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(s.config.Method, s.config.URL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %v", err)
	}

	// Set Content-Type header based on config
	req.Header.Set("Content-Type", s.config.ContentType)

	// Apply authentication
	if err := s.applyAuth(req); err != nil {
		return fmt.Errorf("failed to apply auth: %v", err)
	}

	// Apply custom headers (after auth, so custom headers can override)
	for k, v := range s.config.Headers {
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook notification: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook notification failed with status: %d", resp.StatusCode)
	}
	return nil
}
```

- [ ] **Step 4: Add applyAuth helper method**

Add a new method `applyAuth` on `WebhookSender` right after the `Send` method.

```go
func (s *WebhookSender) applyAuth(req *http.Request) error {
	switch s.config.AuthType {
	case "none":
		return nil
	case "basic":
		var bc BasicAuthConfig
		if err := json.Unmarshal(s.config.AuthConfig, &bc); err != nil {
			return fmt.Errorf("invalid basic auth config: %v", err)
		}
		if bc.Username == "" {
			return fmt.Errorf("basic auth username is required")
		}
		req.SetBasicAuth(bc.Username, bc.Password)
		return nil
	case "custom":
		var cc CustomAuthConfig
		if err := json.Unmarshal(s.config.AuthConfig, &cc); err != nil {
			return fmt.Errorf("invalid custom auth config: %v", err)
		}
		if cc.HeaderName == "" {
			return fmt.Errorf("custom auth header_name is required")
		}
		req.Header.Set(cc.HeaderName, cc.HeaderValue)
		return nil
	default:
		return fmt.Errorf("unsupported auth_type: %s", s.config.AuthType)
	}
}
```

- [ ] **Step 5: Run existing tests to verify backward compatibility**

Run: `go test ./internal/notifier/... -v`
Expected: All existing tests pass (no behavior change for channels without new config fields).

- [ ] **Step 6: Commit**

```bash
git add internal/notifier/webhook.go
git commit -m "feat: extend webhook sender with content_type and auth support"
```

---

### Task 2: Backend — Add WebhookSender Unit Tests

**Files:**
- Modify: `internal/notifier/notifier_test.go`

- [ ] **Step 1: Write test for form-urlencoded content type**

Add a test that uses httptest.Server to verify the webhook sender sends form-urlencoded body correctly.

```go
func TestWebhookSender_FormUrlencoded(t *testing.T) {
	var receivedBody string
	var receivedContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := json.RawMessage(fmt.Sprintf(`{
		"url": "%s",
		"method": "POST",
		"content_type": "application/x-www-form-urlencoded",
		"auth_type": "none"
	}`, server.URL))

	sender, err := NewWebhookSender(config)
	assert.NoError(t, err)

	err = sender.Send("test title", "teams=ops&title=alert&app_content=server down")
	assert.NoError(t, err)
	assert.Equal(t, "application/x-www-form-urlencoded", receivedContentType)
	assert.Equal(t, "teams=ops&title=alert&app_content=server down", receivedBody)
}
```

- [ ] **Step 2: Write test for Basic Auth**

```go
func TestWebhookSender_BasicAuth(t *testing.T) {
	var receivedUser, receivedPass string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUser, receivedPass, _ = r.BasicAuth()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := json.RawMessage(fmt.Sprintf(`{
		"url": "%s",
		"method": "POST",
		"content_type": "application/json",
		"auth_type": "basic",
		"auth_config": {"username": "app_key", "password": "appsecret"}
	}`, server.URL))

	sender, err := NewWebhookSender(config)
	assert.NoError(t, err)

	err = sender.Send("title", "content")
	assert.NoError(t, err)
	assert.Equal(t, "app_key", receivedUser)
	assert.Equal(t, "appsecret", receivedPass)
}
```

- [ ] **Step 3: Write test for Custom Auth**

```go
func TestWebhookSender_CustomAuth(t *testing.T) {
	var receivedHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Custom-Token")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := json.RawMessage(fmt.Sprintf(`{
		"url": "%s",
		"method": "POST",
		"content_type": "application/json",
		"auth_type": "custom",
		"auth_config": {"header_name": "X-Custom-Token", "header_value": "my-token-123"}
	}`, server.URL))

	sender, err := NewWebhookSender(config)
	assert.NoError(t, err)

	err = sender.Send("title", "content")
	assert.NoError(t, err)
	assert.Equal(t, "my-token-123", receivedHeader)
}
```

- [ ] **Step 4: Write test for backward compatibility (no new fields)**

```go
func TestWebhookSender_BackwardCompat_NoNewFields(t *testing.T) {
	var receivedContentType string
	var receivedBody map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := json.RawMessage(fmt.Sprintf(`{
		"url": "%s"
	}`, server.URL))

	sender, err := NewWebhookSender(config)
	assert.NoError(t, err)

	err = sender.Send("hello", "world")
	assert.NoError(t, err)
	assert.Equal(t, "application/json", receivedContentType)
	assert.Equal(t, "hello", receivedBody["title"])
	assert.Equal(t, "world", receivedBody["content"])
}
```

- [ ] **Step 5: Write test for invalid auth_type**

```go
func TestWebhookSender_InvalidAuthType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := json.RawMessage(fmt.Sprintf(`{
		"url": "%s",
		"auth_type": "oauth"
	}`, server.URL))

	sender, err := NewWebhookSender(config)
	assert.NoError(t, err)

	err = sender.Send("title", "content")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported auth_type")
}
```

- [ ] **Step 6: Run all notifier tests**

Run: `go test ./internal/notifier/... -v`
Expected: All tests pass, including new ones.

- [ ] **Step 7: Commit**

```bash
git add internal/notifier/notifier_test.go
git commit -m "test: add webhook sender tests for form-urlencoded and auth"
```

---

### Task 3: Backend — Add Channel Validation for Webhook Config

**Files:**
- Modify: `internal/models/models.go` (lines 82-98, Channel.Validate method)

- [ ] **Step 1: Add webhook config validation in Channel.Validate**

After the existing `validTypes` check (line 94), add a validation block for webhook type config fields. Insert after `if !validTypes[c.Type] { return errors.New("invalid channel type") }` and before `return nil`.

```go
if c.Type == "webhook" {
	var cfg struct {
		Method      string `json:"method"`
		ContentType string `json:"content_type"`
		AuthType    string `json:"auth_type"`
		AuthConfig  json.RawMessage `json:"auth_config"`
	}
	if err := json.Unmarshal(c.Config, &cfg); err == nil {
		validMethods := map[string]bool{"": true, "POST": true, "PUT": true}
		if !validMethods[cfg.Method] {
			return errors.New("invalid webhook method, must be POST or PUT")
		}
		validContentTypes := map[string]bool{"": true, "application/json": true, "application/x-www-form-urlencoded": true}
		if !validContentTypes[cfg.ContentType] {
			return errors.New("invalid webhook content_type")
		}
		validAuthTypes := map[string]bool{"": true, "none": true, "basic": true, "custom": true}
		if !validAuthTypes[cfg.AuthType] {
			return errors.New("invalid webhook auth_type, must be none, basic, or custom")
		}
		if cfg.AuthType == "basic" {
			var bc struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			if err := json.Unmarshal(cfg.AuthConfig, &bc); err != nil || bc.Username == "" {
				return errors.New("basic auth requires username in auth_config")
			}
		}
		if cfg.AuthType == "custom" {
			var cc struct {
				HeaderName string `json:"header_name"`
			}
			if err := json.Unmarshal(cfg.AuthConfig, &cc); err != nil || cc.HeaderName == "" {
				return errors.New("custom auth requires header_name in auth_config")
			}
		}
	}
}
```

Also add `"encoding/json"` to the imports in `models.go`.

- [ ] **Step 2: Run model tests**

Run: `go test ./internal/models/... -v`
Expected: All existing tests pass.

- [ ] **Step 3: Commit**

```bash
git add internal/models/models.go
git commit -m "feat: add webhook config validation for content_type and auth fields"
```

---

### Task 4: Backend — Update maskChannelConfig for Auth Secrets

**Files:**
- Modify: `internal/handlers/config.go` (lines 640-650, maskChannelConfig function)

- [ ] **Step 1: Update maskChannelConfig to mask password and header_value**

The current function masks configs containing `webhook_url`, `secret`, or `sign_key`. Add `password` and `header_value` to the list of sensitive keys.

Replace the function body:

```go
func maskChannelConfig(chType string, config []byte) []byte {
	if config == nil {
		return []byte(`{}`)
	}
	configStr := string(config)
	sensitiveKeys := []string{"webhook_url", "secret", "sign_key", "password", "header_value"}
	for _, key := range sensitiveKeys {
		if strings.Contains(configStr, key) {
			return []byte(`{"masked": true}`)
		}
	}
	return config
}
```

- [ ] **Step 2: Run handler tests**

Run: `go test ./internal/handlers/... -v`
Expected: All existing tests pass.

- [ ] **Step 3: Commit**

```bash
git add internal/handlers/config.go
git commit -m "fix: mask password and header_value in channel config responses"
```

---

### Task 5: Frontend — Extend Channel Type and Config Interface

**Files:**
- Modify: `frontend/src/types/index.ts` (line 72-80, Channel interface)

- [ ] **Step 1: Extend Channel type with webhook config fields**

Replace the Channel interface (lines 72-80):

```typescript
export interface WebhookAuthConfig {
  username?: string
  password?: string
  header_name?: string
  header_value?: string
}

export interface Channel {
  id: number
  name: string
  type: 'feishu' | 'dingtalk' | 'wecom' | 'webhook'
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
  }
  enabled: boolean
  created_at: string
  updated_at: string
}
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/types/index.ts
git commit -m "feat: extend Channel type with webhook auth and content_type fields"
```

---

### Task 6: Frontend — Update Channels.tsx Form

**Files:**
- Modify: `frontend/src/pages/Channels.tsx`

- [ ] **Step 1: Extend ChannelFormValues interface**

Replace the `ChannelFormValues` interface (lines 21-34) to include new webhook config fields:

```typescript
interface ChannelFormValues {
  id?: number
  name: string
  type: string
  enabled: boolean
  config: {
    webhook_url?: string
    secret?: string
    url?: string
    method?: string
    content_type?: string
    headers?: string
    template?: string
    auth_type?: string
    auth_config?: {
      username?: string
      password?: string
      header_name?: string
      header_value?: string
    }
  }
}
```

- [ ] **Step 2: Update formatChannelConfigForForm for webhook type**

Replace the webhook branch in `formatChannelConfigForForm` (lines 75-88):

```typescript
if (channel.type === 'webhook') {
  const authConfig = config.auth_config ?? {}
  return {
    ...channel,
    config: {
      url: String(config.url ?? ''),
      method: String(config.method ?? 'POST'),
      content_type: String(config.content_type ?? 'application/json'),
      headers:
        typeof config.headers === 'string'
          ? config.headers
          : JSON.stringify(config.headers ?? {}, null, 2),
      template: String(config.template ?? ''),
      auth_type: String(config.auth_type ?? 'none'),
      auth_config: {
        username: String(authConfig.username ?? ''),
        password: String(authConfig.password ?? ''),
        header_name: String(authConfig.header_name ?? ''),
        header_value: String(authConfig.header_value ?? ''),
      },
    },
  }
}
```

- [ ] **Step 3: Update buildChannelPayload for webhook type**

Replace the webhook branch in `buildChannelPayload` (lines 96-112):

```typescript
if (values.type === 'webhook') {
  let headers: Record<string, string> = {}
  if (typeof config.headers === 'string' && config.headers.trim()) {
    headers = JSON.parse(config.headers) as Record<string, string>
  }

  const authConfig = config.auth_config ?? {}
  const authPayload: Record<string, string> = {}
  if (values.config.auth_type === 'basic') {
    authPayload.username = String(authConfig.username ?? '')
    authPayload.password = String(authConfig.password ?? '')
  } else if (values.config.auth_type === 'custom') {
    authPayload.header_name = String(authConfig.header_name ?? '')
    authPayload.header_value = String(authConfig.header_value ?? '')
  }

  return {
    ...values,
    config: {
      url: config.url ?? '',
      method: config.method ?? 'POST',
      content_type: config.content_type ?? 'application/json',
      headers,
      template: config.template ?? '',
      auth_type: config.auth_type ?? 'none',
      auth_config: authPayload,
    },
  }
}
```

- [ ] **Step 4: Add new dropdown options**

Add `contentTypeOptions` and `authTypeOptions` arrays after `methodOptions` (after line 170):

```typescript
const contentTypeOptions = [
  { label: 'JSON (application/json)', value: 'application/json' },
  { label: 'Form (application/x-www-form-urlencoded)', value: 'application/x-www-form-urlencoded' },
]

const authTypeOptions = [
  { label: '无认证', value: 'none' },
  { label: 'Basic Auth', value: 'basic' },
  { label: '自定义 Header', value: 'custom' },
]
```

- [ ] **Step 5: Update renderConfigFields for webhook type**

Replace the entire webhook branch in `renderConfigFields` (lines 453-516). This is the biggest change — add content_type dropdown, auth_type dropdown, and conditional auth fields.

```typescript
if (type === 'webhook') {
  return (
    <>
      <div className="flex flex-column gap-2">
        <label className="text-sm">请求 URL</label>
        <InputText
          placeholder="https://example.com/webhook"
          value={formValues.config.url || ''}
          onChange={(e) =>
            setFormValues({
              ...formValues,
              config: { ...formValues.config, url: e.target.value },
            })
          }
          disabled={!canManageConfig}
        />
      </div>
      <div className="flex flex-column gap-2">
        <label className="text-sm">请求方法</label>
        <Dropdown
          value={formValues.config.method || 'POST'}
          options={methodOptions}
          onChange={(e) =>
            setFormValues({
              ...formValues,
              config: { ...formValues.config, method: e.value },
            })
          }
          disabled={!canManageConfig}
        />
      </div>
      <div className="flex flex-column gap-2">
        <label className="text-sm">请求体格式</label>
        <Dropdown
          value={formValues.config.content_type || 'application/json'}
          options={contentTypeOptions}
          onChange={(e) =>
            setFormValues({
              ...formValues,
              config: { ...formValues.config, content_type: e.value },
            })
          }
          disabled={!canManageConfig}
        />
      </div>
      <div className="flex flex-column gap-2">
        <label className="text-sm">请求头（JSON）</label>
        <InputTextarea
          rows={2}
          placeholder='{"Content-Type": "application/json"}'
          value={formValues.config.headers || ''}
          onChange={(e) =>
            setFormValues({
              ...formValues,
              config: { ...formValues.config, headers: e.target.value },
            })
          }
          disabled={!canManageConfig}
        />
      </div>
      <div className="flex flex-column gap-2">
        <label className="text-sm">请求体模板</label>
        <InputTextarea
          rows={3}
          placeholder={
            formValues.config.content_type === 'application/x-www-form-urlencoded'
              ? 'teams=ops&title={{.alert_name}}&app_content={{.message}}'
              : '{"text": "{{.message}}"}'
          }
          value={formValues.config.template || ''}
          onChange={(e) =>
            setFormValues({
              ...formValues,
              config: { ...formValues.config, template: e.target.value },
            })
          }
          disabled={!canManageConfig}
        />
      </div>
      <Divider align="center">
        <span className="text-sm">认证配置</span>
      </Divider>
      <div className="flex flex-column gap-2">
        <label className="text-sm">认证方式</label>
        <Dropdown
          value={formValues.config.auth_type || 'none'}
          options={authTypeOptions}
          onChange={(e) =>
            setFormValues({
              ...formValues,
              config: {
                ...formValues.config,
                auth_type: e.value,
                auth_config: {},
              },
            })
          }
          disabled={!canManageConfig}
        />
      </div>
      {formValues.config.auth_type === 'basic' && (
        <>
          <div className="flex flex-column gap-2">
            <label className="text-sm">用户名</label>
            <InputText
              placeholder="app_key"
              value={formValues.config.auth_config?.username || ''}
              onChange={(e) =>
                setFormValues({
                  ...formValues,
                  config: {
                    ...formValues.config,
                    auth_config: { ...formValues.config.auth_config, username: e.target.value },
                  },
                })
              }
              disabled={!canManageConfig}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="text-sm">密码</label>
            <InputText
              placeholder="appsecret"
              value={formValues.config.auth_config?.password || ''}
              onChange={(e) =>
                setFormValues({
                  ...formValues,
                  config: {
                    ...formValues.config,
                    auth_config: { ...formValues.config.auth_config, password: e.target.value },
                  },
                })
              }
              disabled={!canManageConfig}
            />
          </div>
        </>
      )}
      {formValues.config.auth_type === 'custom' && (
        <>
          <div className="flex flex-column gap-2">
            <label className="text-sm">Header 名称</label>
            <InputText
              placeholder="Authorization"
              value={formValues.config.auth_config?.header_name || ''}
              onChange={(e) =>
                setFormValues({
                  ...formValues,
                  config: {
                    ...formValues.config,
                    auth_config: { ...formValues.config.auth_config, header_name: e.target.value },
                  },
                })
              }
              disabled={!canManageConfig}
            />
          </div>
          <div className="flex flex-column gap-2">
            <label className="text-sm">Header 值</label>
            <InputText
              placeholder="Bearer xxx"
              value={formValues.config.auth_config?.header_value || ''}
              onChange={(e) =>
                setFormValues({
                  ...formValues,
                  config: {
                    ...formValues.config,
                    auth_config: { ...formValues.config.auth_config, header_value: e.target.value },
                  },
                })
              }
              disabled={!canManageConfig}
            />
          </div>
        </>
      )}
    </>
  )
}
```

- [ ] **Step 6: Run frontend build to verify no type errors**

Run: `cd frontend && pnpm build`
Expected: Build succeeds with no TypeScript errors.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/pages/Channels.tsx
git commit -m "feat: add webhook content_type and auth config fields to channel form"
```

---

### Task 7: Frontend — Run Full Test Suite

**Files:** None (verification only)

- [ ] **Step 1: Run frontend lint**

Run: `cd frontend && pnpm lint`
Expected: No lint errors.

- [ ] **Step 2: Run frontend tests**

Run: `cd frontend && pnpm test -- --run`
Expected: All tests pass.

- [ ] **Step 3: Run backend tests**

Run: `go test ./...`
Expected: All tests pass.

---

## Self-Review Checklist

**1. Spec coverage:**
- Config JSON extension (content_type, auth_type, auth_config) → Task 1
- Sender logic for form-urlencoded → Task 1
- Sender logic for Basic Auth → Task 1
- Sender logic for Custom Auth → Task 1
- Channel validation for new fields → Task 3
- maskChannelConfig for password/header_value → Task 4
- Frontend type extension → Task 5
- Frontend form fields → Task 6
- Backward compatibility → Task 2 (test)

**2. Placeholder scan:** No TBD/TODO found. All steps contain complete code.

**3. Type consistency:**
- `WebhookConfig.AuthConfig` is `json.RawMessage` in Go — matches `BasicAuthConfig`/`CustomAuthConfig` unmarshal targets
- `ChannelFormValues.config.auth_config` is `{username?, password?, header_name?, header_value?}` in TypeScript — matches `WebhookAuthConfig` in types/index.ts
- `buildChannelPayload` produces `auth_config` as `Record<string, string>` — matches Go's `json.RawMessage` (JSON object with string values)