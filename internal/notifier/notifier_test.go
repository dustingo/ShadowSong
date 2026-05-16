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
