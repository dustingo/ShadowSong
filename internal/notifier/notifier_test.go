package notifier

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func TestSendToChannel_UnsupportedTypeIncludesChannelContext(t *testing.T) {
	channel := &models.Channel{
		ID:     7,
		Name:   "broken-channel",
		Type:   "unknown",
		Config: datatypes.JSON(`{}`),
	}

	err := SendToChannel(channel, "title", "content", nil)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "channel 7 (broken-channel)")
		assert.Contains(t, err.Error(), "unsupported type")
	}
}

func TestIsRetryableSendError_TransientSendFailuresAreRetryable(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{
			name: "transport failure",
			err:  errors.New("channel 11 (ops-webhook) send failed: failed to send webhook notification: dial tcp timeout"),
		},
		{
			name: "upstream service unavailable",
			err:  errors.New("channel 11 (ops-webhook) send failed: webhook notification failed with status: 503"),
		},
		{
			name: "feishu upstream rate limit",
			err:  errors.New("channel 11 (ops-webhook) send failed: feishu notification failed, status: 429, body: busy"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.True(t, IsRetryableSendError(tc.err))
		})
	}
}

func TestIsRetryableSendError_DeterministicFailuresRemainTerminal(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{
			name: "unsupported channel type",
			err:  errors.New("channel 7 (broken-channel) unsupported type: unknown"),
		},
		{
			name: "sender init failure",
			err:  errors.New("channel 9 (ops-feishu) sender init failed: feishu webhook_url is required"),
		},
		{
			name: "template render failure",
			err:  errors.New("template execute error: map has no entry for key \"labels\""),
		},
		{
			name: "datasource lookup failure",
			err:  errors.New("data source not found for source=prometheus"),
		},
		{
			name: "default notification fallback init failure",
			err:  errors.New("channel 10 (ops-webhook) send failed: failed to create webhook request: parse \"::\": missing protocol scheme"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.False(t, IsRetryableSendError(tc.err))
		})
	}
}

func TestWebhookSender_FormUrlencoded(t *testing.T) {
	var receivedContentType string
	var receivedBody string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := fmt.Sprintf(`{"url":"%s","content_type":"application/x-www-form-urlencoded"}`, ts.URL)
	sender, err := NewWebhookSender(json.RawMessage(config))
	assert.NoError(t, err)

	err = sender.Send("alert-title", "alert-content", nil)
	assert.NoError(t, err)

	assert.Equal(t, "application/x-www-form-urlencoded", receivedContentType)
	assert.Equal(t, "alert-content", receivedBody)
}

func TestWebhookSender_BasicAuth(t *testing.T) {
	var receivedUser, receivedPass string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUser, receivedPass, _ = r.BasicAuth()
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := fmt.Sprintf(`{"url":"%s","auth_type":"basic","auth_config":{"username":"myuser","password":"mypass"}}`, ts.URL)
	sender, err := NewWebhookSender(json.RawMessage(config))
	assert.NoError(t, err)

	err = sender.Send("title", "content", nil)
	assert.NoError(t, err)

	assert.Equal(t, "myuser", receivedUser)
	assert.Equal(t, "mypass", receivedPass)
}

func TestWebhookSender_CustomAuth(t *testing.T) {
	var receivedHeader string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Custom-Token")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := fmt.Sprintf(`{"url":"%s","auth_type":"custom","auth_config":{"header_name":"X-Custom-Token","header_value":"secret-token-123"}}`, ts.URL)
	sender, err := NewWebhookSender(json.RawMessage(config))
	assert.NoError(t, err)

	err = sender.Send("title", "content", nil)
	assert.NoError(t, err)

	assert.Equal(t, "secret-token-123", receivedHeader)
}

func TestWebhookSender_BackwardCompat_NoNewFields(t *testing.T) {
	var receivedContentType string
	var receivedBody string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := fmt.Sprintf(`{"url":"%s"}`, ts.URL)
	sender, err := NewWebhookSender(json.RawMessage(config))
	assert.NoError(t, err)

	err = sender.Send("my-title", "my-content", nil)
	assert.NoError(t, err)

	assert.Equal(t, "application/json", receivedContentType)

	var payload map[string]string
	assert.NoError(t, json.Unmarshal([]byte(receivedBody), &payload))
	assert.Equal(t, "my-title", payload["title"])
	assert.Equal(t, "my-content", payload["content"])
}

func TestWebhookSender_InvalidAuthType(t *testing.T) {
	var received bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Bypass constructor validation to test applyAuth's default branch
	sender := &WebhookSender{
		config: WebhookConfig{
			URL:         ts.URL,
			Method:      "POST",
			ContentType: "application/json",
			AuthType:    "oauth",
		},
		client: &http.Client{Timeout: 10 * time.Second},
	}

	err := sender.Send("title", "content", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported auth_type")
	assert.False(t, received, "request should not have been sent")
}

func TestWebhookSender_InvalidMethod(t *testing.T) {
	config := `{"url":"http://example.com","method":"DELETE"}`
	sender, err := NewWebhookSender(json.RawMessage(config))
	assert.Error(t, err)
	assert.Nil(t, sender)
	assert.Contains(t, err.Error(), "webhook method must be POST or PUT")
}

func TestWebhookSender_InvalidContentType(t *testing.T) {
	config := `{"url":"http://example.com","content_type":"text/plain"}`
	sender, err := NewWebhookSender(json.RawMessage(config))
	assert.Error(t, err)
	assert.Nil(t, sender)
	assert.Contains(t, err.Error(), "webhook content_type must be application/json or application/x-www-form-urlencoded")
}

func TestWebhookSender_BasicAuthMissingUsername(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	config := fmt.Sprintf(`{"url":"%s","auth_type":"basic","auth_config":{}}`, ts.URL)
	sender, err := NewWebhookSender(json.RawMessage(config))
	assert.NoError(t, err)

	err = sender.Send("title", "content", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "basic auth requires username")
}

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
		"title":      "[P1] DiskFull",
		"content":    "磁盘使用率99%",
		"alert_name": "DiskFull",
		"severity":   "P1",
		"message":    "磁盘使用率99%",
	}
	err = sender.Send("[P1] DiskFull", "磁盘使用率99%", data)
	assert.NoError(t, err)

	var payload map[string]string
	assert.NoError(t, json.Unmarshal([]byte(receivedBody), &payload))
	assert.Equal(t, "DiskFull", payload["alert"])
	assert.Equal(t, "P1", payload["level"])
	assert.Equal(t, "磁盘使用率99%", payload["msg"])
}

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
