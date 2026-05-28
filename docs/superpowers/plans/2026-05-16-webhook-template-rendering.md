# Channel Webhook 请求体模板渲染 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 Channel webhook 的请求体模板支持 Go 模板渲染，能引用 OutputTemplate 渲染结果 (title/content) 和告警原始字段。

**Architecture:** 修改 Sender 接口签名新增 data 参数，WebhookSender.Send 内部用 template.Renderer 渲染请求体模板，调用链从 webhook handler 透传 data 上下文到 notifier。

**Tech Stack:** Go, text/template, existing template.Renderer

---

### Task 1: 修改 Sender 接口签名

**Files:**
- Modify: `internal/notifier/notifier.go:16-18`
- Test: `internal/notifier/notifier_test.go`

- [ ] **Step 1: Write failing test for new Send signature**

Add a test that calls `SendToChannel` with the new signature (4 args instead of 3). This will fail to compile.

```go
func TestSendToChannel_NewSignature(t *testing.T) {
	channel := &models.Channel{
		ID:     1,
		Name:   "test",
		Type:   "feishu",
		Config: datatypes.JSON(`{"webhook_url":"http://example.com"}`),
	}
	// New signature: SendToChannel(channel, title, content, data)
	err := SendToChannel(channel, "title", "content", map[string]interface{}{"alert_name": "test"})
	// We just need this to compile — the actual behavior is tested per-sender
	_ = err
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/notifier/ -run TestSendToChannel_NewSignature -v`
Expected: compile error — `SendToChannel` has wrong number of args

- [ ] **Step 3: Update Sender interface and SendToChannel**

```go
type Sender interface {
	Send(title, content string, data map[string]interface{}) error
}
```

Update `SendToChannel`:

```go
func SendToChannel(channel *models.Channel, title, content string, data map[string]interface{}) error {
	var sender Sender
	var err error

	configBytes := []byte(channel.Config)

	switch channel.Type {
	case "feishu":
		sender, err = NewFeishuSender(configBytes)
	case "dingtalk":
		sender, err = NewDingTalkSender(configBytes)
	case "wecom":
		sender, err = NewWeComSender(configBytes)
	case "webhook":
		sender, err = NewWebhookSender(configBytes)
	default:
		return fmt.Errorf("channel %d (%s) unsupported type: %s", channel.ID, channel.Name, channel.Type)
	}

	if err != nil {
		return fmt.Errorf("channel %d (%s) sender init failed: %w", channel.ID, channel.Name, err)
	}

	if err := sender.Send(title, content, data); err != nil {
		return fmt.Errorf("channel %d (%s) send failed: %w", channel.ID, channel.Name, err)
	}

	return nil
}
```

- [ ] **Step 4: Update all Sender implementations to match new interface**

FeishuSender.Send:
```go
func (s *FeishuSender) Send(title, content string, data map[string]interface{}) error {
```

DingTalkSender.Send:
```go
func (s *DingTalkSender) Send(title, content string, data map[string]interface{}) error {
```

WeComSender.Send:
```go
func (s *WeComSender) Send(title, content string, data map[string]interface{}) error {
```

WebhookSender.Send:
```go
func (s *WebhookSender) Send(title, content string, data map[string]interface{}) error {
```

All three non-webhook senders ignore `data` — no logic change needed in their bodies.

- [ ] **Step 5: Update all existing test calls to use new signature**

In `internal/notifier/notifier_test.go`, update every `Send` and `SendToChannel` call to pass `nil` as the 4th argument:

- `sender.Send("alert-title", "alert-content")` → `sender.Send("alert-title", "alert-content", nil)`
- `sender.Send("title", "content")` → `sender.Send("title", "content", nil)`
- `sender.Send("my-title", "my-content")` → `sender.Send("my-title", "my-content", nil)`
- `SendToChannel(channel, "title", "content")` → `SendToChannel(channel, "title", "content", nil)`

- [ ] **Step 6: Run all notifier tests to verify they pass**

Run: `go test ./internal/notifier/ -v`
Expected: all PASS

- [ ] **Step 7: Commit**

```bash
git add internal/notifier/notifier.go internal/notifier/notifier_test.go
git commit -m "refactor: add data parameter to Sender interface for template rendering"
```

---

### Task 2: WebhookSender 模板渲染逻辑

**Files:**
- Modify: `internal/notifier/notifier.go:326-409`
- Test: `internal/notifier/notifier_test.go`

- [ ] **Step 1: Write failing test — JSON template with title/content**

```go
func TestWebhookSender_TemplateRendersTitleContent(t *testing.T) {
	var receivedBody string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := fmt.Sprintf(`{"url":"%s","template":"{\"msg_type\":\"alert\",\"text\":\"{{.content}}\",\"summary\":\"{{.title}}\"}"}`, ts.URL)
	sender, err := NewWebhookSender(json.RawMessage(config))
	assert.NoError(t, err)

	data := map[string]interface{}{
		"title":   "[P0] CPU过高",
		"content": "CPU使用率95%",
	}
	err = sender.Send("[P0] CPU过高", "CPU使用率95%", data)
	assert.NoError(t, err)

	var payload map[string]string
	assert.NoError(t, json.Unmarshal([]byte(receivedBody), &payload))
	assert.Equal(t, "alert", payload["msg_type"])
	assert.Equal(t, "CPU使用率95%", payload["text"])
	assert.Equal(t, "[P0] CPU过高", payload["summary"])
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/notifier/ -run TestWebhookSender_TemplateRendersTitleContent -v`
Expected: FAIL — template is sent as raw string, `{{.content}}` appears literally

- [ ] **Step 3: Implement template rendering in WebhookSender**

Add `renderer` field to WebhookSender and update `NewWebhookSender`:

```go
type WebhookSender struct {
	config   WebhookConfig
	client   *http.Client
	renderer *template.Renderer
}
```

Add import at top of file:
```go
tmpl "github.com/game-ops/ai-alert-system/internal/template"
```

Update `NewWebhookSender` to create renderer:
```go
return &WebhookSender{
	config:   wc,
	client:   &http.Client{Timeout: 10 * time.Second},
	renderer: tmpl.NewRenderer(),
}, nil
```

Update `Send` method — replace the Template branch:

```go
func (s *WebhookSender) Send(title, content string, data map[string]interface{}) error {
	var reqBody io.Reader

	switch s.config.ContentType {
	case "application/x-www-form-urlencoded":
		if s.config.Template != "" {
			rendered, err := s.renderer.Render(s.config.Template, data)
			if err != nil {
				return fmt.Errorf("failed to render webhook template: %w", err)
			}
			reqBody = strings.NewReader(rendered)
		} else {
			reqBody = strings.NewReader(content)
		}
	default:
		// JSON content type
		var body []byte
		if s.config.Template != "" {
			rendered, err := s.renderer.Render(s.config.Template, data)
			if err != nil {
				return fmt.Errorf("failed to render webhook template: %w", err)
			}
			body = []byte(rendered)
		} else {
			payload := map[string]string{
				"title":   title,
				"content": content,
			}
			body, _ = json.Marshal(payload)
		}
		reqBody = bytes.NewReader(body)
	}

	// ... rest unchanged
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/notifier/ -run TestWebhookSender_TemplateRendersTitleContent -v`
Expected: PASS

- [ ] **Step 5: Write failing test — template with alert original fields**

```go
func TestWebhookSender_TemplateRendersAlertFields(t *testing.T) {
	var receivedBody string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := fmt.Sprintf(`{"url":"%s","template":"{\"alert\":\"{{.alert_name}}\",\"level\":\"{{.severity}}\",\"msg\":\"{{.message}}\"}"}`, ts.URL)
	sender, err := NewWebhookSender(json.RawMessage(config))
	assert.NoError(t, err)

	data := map[string]interface{}{
		"title":      "[P1] 磁盘满",
		"content":    "磁盘使用率99%",
		"alert_name": "DiskFull",
		"severity":   "P1",
		"message":    "磁盘使用率99%",
	}
	err = sender.Send("[P1] 磁盘满", "磁盘使用率99%", data)
	assert.NoError(t, err)

	var payload map[string]string
	assert.NoError(t, json.Unmarshal([]byte(receivedBody), &payload))
	assert.Equal(t, "DiskFull", payload["alert"])
	assert.Equal(t, "P1", payload["level"])
	assert.Equal(t, "磁盘使用率99%", payload["msg"])
}
```

- [ ] **Step 6: Run test to verify it passes**

Run: `go test ./internal/notifier/ -run TestWebhookSender_TemplateRendersAlertFields -v`
Expected: PASS (renderer already supports these fields)

- [ ] **Step 7: Write failing test — template with event field**

```go
func TestWebhookSender_TemplateRendersEventField(t *testing.T) {
	var receivedBody string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := fmt.Sprintf(`{"url":"%s","template":"{\"host\":\"{{.event.host}}\",\"metric\":\"{{.event.metric}}\"}"}`, ts.URL)
	sender, err := NewWebhookSender(json.RawMessage(config))
	assert.NoError(t, err)

	data := map[string]interface{}{
		"title":   "test",
		"content": "test content",
		"event": map[string]interface{}{
			"host":   "server-01",
			"metric": "cpu_usage",
		},
	}
	err = sender.Send("test", "test content", data)
	assert.NoError(t, err)

	var payload map[string]string
	assert.NoError(t, json.Unmarshal([]byte(receivedBody), &payload))
	assert.Equal(t, "server-01", payload["host"])
	assert.Equal(t, "cpu_usage", payload["metric"])
}
```

- [ ] **Step 8: Run test to verify it passes**

Run: `go test ./internal/notifier/ -run TestWebhookSender_TemplateRendersEventField -v`
Expected: PASS

- [ ] **Step 9: Write failing test — template with labels field**

```go
func TestWebhookSender_TemplateRendersLabels(t *testing.T) {
	var receivedBody string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := fmt.Sprintf(`{"url":"%s","template":"{\"team\":\"{{.labels.team}}\",\"env\":\"{{.labels.env}}\"}"}`, ts.URL)
	sender, err := NewWebhookSender(json.RawMessage(config))
	assert.NoError(t, err)

	data := map[string]interface{}{
		"title":   "test",
		"content": "test content",
		"labels": map[string]interface{}{
			"team": "ops",
			"env":  "prod",
		},
	}
	err = sender.Send("test", "test content", data)
	assert.NoError(t, err)

	var payload map[string]string
	assert.NoError(t, json.Unmarshal([]byte(receivedBody), &payload))
	assert.Equal(t, "ops", payload["team"])
	assert.Equal(t, "prod", payload["env"])
}
```

- [ ] **Step 10: Run test to verify it passes**

Run: `go test ./internal/notifier/ -run TestWebhookSender_TemplateRendersLabels -v`
Expected: PASS

- [ ] **Step 11: Write failing test — invalid template syntax falls back**

```go
func TestWebhookSender_InvalidTemplateReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := fmt.Sprintf(`{"url":"%s","template":"{{.broken"}`, ts.URL)
	sender, err := NewWebhookSender(json.RawMessage(config))
	assert.NoError(t, err)

	data := map[string]interface{}{
		"title":   "test",
		"content": "test content",
	}
	err = sender.Send("test", "test content", data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to render webhook template")
}
```

- [ ] **Step 12: Run test to verify it passes**

Run: `go test ./internal/notifier/ -run TestWebhookSender_InvalidTemplateReturnsError -v`
Expected: PASS

- [ ] **Step 13: Write failing test — form-urlencoded with template**

```go
func TestWebhookSender_FormUrlencodedWithTemplate(t *testing.T) {
	var receivedBody string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := fmt.Sprintf(`{"url":"%s","content_type":"application/x-www-form-urlencoded","template":"title={{.title}}&content={{.content}}"}`, ts.URL)
	sender, err := NewWebhookSender(json.RawMessage(config))
	assert.NoError(t, err)

	data := map[string]interface{}{
		"title":   "AlertTitle",
		"content": "AlertContent",
	}
	err = sender.Send("AlertTitle", "AlertContent", data)
	assert.NoError(t, err)

	assert.Equal(t, "title=AlertTitle&content=AlertContent", receivedBody)
}
```

- [ ] **Step 14: Run test to verify it passes**

Run: `go test ./internal/notifier/ -run TestWebhookSender_FormUrlencodedWithTemplate -v`
Expected: PASS

- [ ] **Step 15: Run all notifier tests**

Run: `go test ./internal/notifier/ -v`
Expected: all PASS

- [ ] **Step 16: Commit**

```bash
git add internal/notifier/notifier.go internal/notifier/notifier_test.go
git commit -m "feat: render webhook channel template with Go template engine"
```

---

### Task 3: 更新 webhook handler 调用链透传 data

**Files:**
- Modify: `internal/handlers/webhook.go:45-46,781-786,1126-1145`
- Test: `internal/handlers/webhook_test.go`

- [ ] **Step 1: Update sendToChannel field type**

Change the field type in `WebhookHandler` struct (line 45):

```go
sendToChannel func(channel *models.Channel, title, content string, data map[string]interface{}) error
```

Update `NewWebhookHandler` (line 56):

```go
sendToChannel:   notifier.SendToChannel,
```

Update `notificationSender` method (line 781-786):

```go
func (h *WebhookHandler) notificationSender() func(channel *models.Channel, title, content string, data map[string]interface{}) error {
	if h.sendToChannel != nil {
		return h.sendToChannel
	}
	return notifier.SendToChannel
}
```

- [ ] **Step 2: Update sendChannelNotification to build data and pass it**

```go
func (h *WebhookHandler) sendChannelNotification(
	alert *models.Alert,
	channel *models.Channel,
	routeRule *models.RouteRule,
	title string,
	content string,
	mode string,
) {
	sender := h.notificationSender()
	deliveryRecord, deliveryErr := h.startNotificationDelivery(alert, channel, routeRule, title, content, mode)
	if deliveryErr != nil {
		h.logAlertEvent("delivery_start", h.eventFields(h.traceFieldsForNotification(alert, channel), map[string]string{
			"mode":  mode,
			"error": deliveryErr.Error(),
		}), "failed to persist delivery envelope")
	}

	data := h.buildNotificationRenderContext(alert)
	data["title"] = title
	data["content"] = content

	for attempt := 1; attempt <= notificationMaxAttempts; attempt++ {
		startedAt := time.Now()
		err := sender(channel, title, content, data)
		attemptFields := h.eventFields(h.traceFieldsForAttempt(alert, channel, attempt, notificationMaxAttempts, err), map[string]string{
			"mode": mode,
		})
		h.recordNotificationAttempt(deliveryRecord, attempt, err, time.Since(startedAt), mode)
		h.logAlertEvent("send_attempt", attemptFields, "notification attempt recorded")

		if err == nil {
			h.markNotificationDelivered(deliveryRecord, attempt)
			h.logAlertEvent("send_notification", attemptFields, "notification sent")
			return
		}

		if !notifier.IsRetryableSendError(err) {
			h.markNotificationFailed(deliveryRecord, attempt, err, false, false)
			h.logAlertEvent("send_notification", attemptFields, "failed to send %s notification", mode)
			return
		}

		if attempt == notificationMaxAttempts {
			h.markNotificationFailed(deliveryRecord, attempt, err, true, true)
			h.logAlertEvent("terminal_failure", attemptFields, "retry budget exhausted for %s notification", mode)
			return
		}

		h.sleepFunc()(50 * time.Millisecond)
	}
}
```

- [ ] **Step 3: Update all test mocks for sendToChannel**

In `internal/handlers/webhook_test.go`, every `sendToChannel` assignment needs the new signature. For each occurrence, change:

```go
sendToChannel: func(channel *models.Channel, title, content string) error {
```

to:

```go
sendToChannel: func(channel *models.Channel, title, content string, data map[string]interface{}) error {
```

- [ ] **Step 4: Update config.go TestChannel handler**

In `internal/handlers/config.go:359`:

```go
if err := notifier.SendToChannel(&ch, testTitle, testContent, nil); err != nil {
```

- [ ] **Step 5: Run all handler tests**

Run: `go test ./internal/handlers/ -v`
Expected: all PASS

- [ ] **Step 6: Commit**

```bash
git add internal/handlers/webhook.go internal/handlers/webhook_test.go internal/handlers/config.go
git commit -m "feat: pass notification render context through to channel sender"
```

---

### Task 4: 端到端验证

**Files:**
- Test: `internal/notifier/notifier_test.go`

- [ ] **Step 1: Write end-to-end test combining OutputTemplate rendering + webhook template**

```go
func TestWebhookSender_FullPipeline(t *testing.T) {
	var receivedBody string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Simulate what the handler builds: data from buildNotificationRenderContext + title/content
	data := map[string]interface{}{
		"title":         "[P0] CPU过高",
		"content":       "严重告警\n名称: CPU过高\n状态: firing",
		"alert_id":      "alert-123",
		"alert_name":    "CPU过高",
		"severity":      "P0",
		"severity_code": "P0",
		"message":       "CPU使用率95%",
		"source":        "prometheus",
		"status":        "firing",
		"labels": map[string]interface{}{
			"team": "ops",
			"env":  "prod",
		},
		"event": map[string]interface{}{
			"host":    "server-01",
			"metric":  "cpu_usage",
			"value":   "95",
		},
		"alert": map[string]interface{}{
			"id":       "alert-123",
			"name":     "CPU过高",
			"severity": "P0",
			"message":  "CPU使用率95%",
			"source":   "prometheus",
			"status":   "firing",
			"labels": map[string]interface{}{
				"team": "ops",
				"env":  "prod",
			},
		},
	}

	config := fmt.Sprintf(`{"url":"%s","template":"{\"msg_type\":\"alert\",\"title\":\"{{.title}}\",\"level\":\"{{.severity}}\",\"host\":\"{{.event.host}}\",\"team\":\"{{.labels.team}}\"}"}`, ts.URL)
	sender, err := NewWebhookSender(json.RawMessage(config))
	assert.NoError(t, err)

	err = sender.Send("[P0] CPU过高", "严重告警...", data)
	assert.NoError(t, err)

	var payload map[string]string
	assert.NoError(t, json.Unmarshal([]byte(receivedBody), &payload))
	assert.Equal(t, "alert", payload["msg_type"])
	assert.Equal(t, "[P0] CPU过高", payload["title"])
	assert.Equal(t, "P0", payload["level"])
	assert.Equal(t, "server-01", payload["host"])
	assert.Equal(t, "ops", payload["team"])
}
```

- [ ] **Step 2: Run test to verify it passes**

Run: `go test ./internal/notifier/ -run TestWebhookSender_FullPipeline -v`
Expected: PASS

- [ ] **Step 3: Run full test suite**

Run: `go test ./... 2>&1 | tail -30`
Expected: all PASS

- [ ] **Step 4: Commit**

```bash
git add internal/notifier/notifier_test.go
git commit -m "test: add end-to-end test for webhook template rendering pipeline"
```