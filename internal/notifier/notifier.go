package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
	tmpl "github.com/game-ops/ai-alert-system/internal/template"
)

type Sender interface {
	Send(title, content string, data map[string]interface{}) error
}

var retryableHTTPStatusCodes = map[int]struct{}{
	http.StatusRequestTimeout:      {},
	http.StatusTooManyRequests:     {},
	http.StatusInternalServerError: {},
	http.StatusBadGateway:          {},
	http.StatusServiceUnavailable:  {},
	http.StatusGatewayTimeout:      {},
}

// SendToChannel sends notification to the specified channel
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

// IsRetryableSendError reports whether a wrapped SendToChannel error is a transient send-stage failure.
func IsRetryableSendError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	if !strings.Contains(message, " send failed: ") {
		return false
	}

	if strings.Contains(message, "failed to create webhook request") ||
		strings.Contains(message, "failed to marshal") ||
		strings.Contains(message, "failed to unmarshal") {
		return false
	}

	if status, ok := extractTrailingStatusCode(message); ok {
		_, retryable := retryableHTTPStatusCodes[status]
		return retryable
	}

	transientMarkers := []string{
		"timeout",
		"tempor",
		"connection reset",
		"connection refused",
		"broken pipe",
		"unexpected eof",
		"eof",
		"no such host",
		"tls handshake timeout",
		"i/o timeout",
		"dial tcp",
	}

	for _, marker := range transientMarkers {
		if strings.Contains(message, marker) {
			return true
		}
	}

	return false
}

func extractTrailingStatusCode(message string) (int, bool) {
	statusMarkers := []string{"status: ", "status code: "}
	for _, marker := range statusMarkers {
		idx := strings.Index(message, marker)
		if idx == -1 {
			continue
		}

		start := idx + len(marker)
		end := start
		for end < len(message) && message[end] >= '0' && message[end] <= '9' {
			end++
		}

		if end == start {
			continue
		}

		status, err := strconv.Atoi(message[start:end])
		if err != nil {
			continue
		}

		return status, true
	}

	return 0, false
}

// ============ Feishu Sender ============

type FeishuConfig struct {
	WebhookURL string `json:"webhook_url"`
	Secret     string `json:"secret"`
}

type FeishuSender struct {
	config FeishuConfig
	client *http.Client
}

func NewFeishuSender(config json.RawMessage) (Sender, error) {
	var fc FeishuConfig
	if err := json.Unmarshal(config, &fc); err != nil {
		return nil, err
	}
	if fc.WebhookURL == "" {
		return nil, fmt.Errorf("feishu webhook_url is required")
	}
	return &FeishuSender{
		config: fc,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

type FeishuMessage struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

func (s *FeishuSender) Send(title, content string, data map[string]interface{}) error {
	msg := FeishuMessage{}
	msg.MsgType = "text"
	msg.Content.Text = fmt.Sprintf("**%s**\n%s", title, content)

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal feishu message: %v", err)
	}
	resp, err := s.client.Post(s.config.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to send feishu notification: %v", err)
	}
	defer resp.Body.Close()
	// 必须读取body连接才能复用
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return fmt.Errorf("failed to read feishu response body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("feishu notification failed, status: %d, body: %s", resp.StatusCode, respBody)
	}
	// 解析飞书业务状态码，因为飞书正常都是返回200
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to unmarshal feishu response body: %v", err)
	}
	if result.Code != 0 {
		return fmt.Errorf("feishu notification failed, code: %d, msg: %s", result.Code, result.Msg)
	}
	return nil
}

// ============ DingTalk Sender ============

type DingTalkConfig struct {
	WebhookURL string `json:"webhook_url"`
	Secret     string `json:"secret"`
}

type DingTalkSender struct {
	config DingTalkConfig
	client *http.Client
}

func NewDingTalkSender(config json.RawMessage) (Sender, error) {
	var dc DingTalkConfig
	if err := json.Unmarshal(config, &dc); err != nil {
		return nil, err
	}
	if dc.WebhookURL == "" {
		return nil, fmt.Errorf("dingtalk webhook_url is required")
	}
	return &DingTalkSender{
		config: dc,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

type DingTalkMessage struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}

func (s *DingTalkSender) Send(title, content string, data map[string]interface{}) error {
	msg := DingTalkMessage{}
	msg.MsgType = "text"
	msg.Text.Content = fmt.Sprintf("%s\n%s", title, content)

	body, _ := json.Marshal(msg)
	resp, err := s.client.Post(s.config.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to send dingtalk notification: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dingtalk notification failed with status: %d", resp.StatusCode)
	}
	return nil
}

// ============ WeCom Sender ============

type WeComConfig struct {
	WebhookURL string `json:"webhook_url"`
}

type WeComSender struct {
	config WeComConfig
	client *http.Client
}

func NewWeComSender(config json.RawMessage) (Sender, error) {
	var wc WeComConfig
	if err := json.Unmarshal(config, &wc); err != nil {
		return nil, err
	}
	if wc.WebhookURL == "" {
		return nil, fmt.Errorf("wecom webhook_url is required")
	}
	return &WeComSender{
		config: wc,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

type WeComMessage struct {
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}

func (s *WeComSender) Send(title, content string, data map[string]interface{}) error {
	msg := WeComMessage{}
	msg.MsgType = "text"
	msg.Text.Content = fmt.Sprintf("%s\n%s", title, content)

	body, _ := json.Marshal(msg)
	resp, err := s.client.Post(s.config.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to send wecom notification: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wecom notification failed with status: %d", resp.StatusCode)
	}
	return nil
}

// ============ Webhook Sender ============

type WebhookConfig struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	Template    string            `json:"template"`
	ContentType string            `json:"content_type"`
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

type WebhookSender struct {
	config   WebhookConfig
	client   *http.Client
	renderer *tmpl.Renderer
}

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
	if wc.Method != "POST" && wc.Method != "PUT" {
		return nil, fmt.Errorf("webhook method must be POST or PUT, got: %s", wc.Method)
	}
	if wc.ContentType == "" {
		wc.ContentType = "application/json"
	}
	if wc.ContentType != "application/json" && wc.ContentType != "application/x-www-form-urlencoded" {
		return nil, fmt.Errorf("webhook content_type must be application/json or application/x-www-form-urlencoded, got: %s", wc.ContentType)
	}
	if wc.AuthType == "" {
		wc.AuthType = "none"
	}
	if wc.AuthType != "none" && wc.AuthType != "basic" && wc.AuthType != "custom" {
		return nil, fmt.Errorf("webhook auth_type must be none, basic, or custom, got: %s", wc.AuthType)
	}
	return &WebhookSender{
		config:   wc,
		client:   &http.Client{Timeout: 10 * time.Second},
		renderer: tmpl.NewRenderer(),
	}, nil
}

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

	req, err := http.NewRequest(s.config.Method, s.config.URL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %v", err)
	}

	req.Header.Set("Content-Type", s.config.ContentType)

	for k, v := range s.config.Headers {
		req.Header.Set(k, v)
	}

	if err := s.applyAuth(req); err != nil {
		return err
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

func (s *WebhookSender) applyAuth(req *http.Request) error {
	switch s.config.AuthType {
	case "none":
		return nil
	case "basic":
		var bac BasicAuthConfig
		if err := json.Unmarshal(s.config.AuthConfig, &bac); err != nil {
			return fmt.Errorf("failed to unmarshal basic auth config: %v", err)
		}
		if bac.Username == "" {
			return fmt.Errorf("basic auth requires username")
		}
		req.SetBasicAuth(bac.Username, bac.Password)
		return nil
	case "custom":
		var cac CustomAuthConfig
		if err := json.Unmarshal(s.config.AuthConfig, &cac); err != nil {
			return fmt.Errorf("failed to unmarshal custom auth config: %v", err)
		}
		if cac.HeaderName == "" {
			return fmt.Errorf("custom auth requires header_name")
		}
		req.Header.Set(cac.HeaderName, cac.HeaderValue)
		return nil
	default:
		return fmt.Errorf("unsupported auth_type: %s", s.config.AuthType)
	}
}
