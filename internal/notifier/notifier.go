package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/game-ops/ai-alert-system/internal/models"
)

type Sender interface {
	Send(title, content string) error
}

// SendToChannel sends notification to the specified channel
func SendToChannel(channel *models.Channel, title, content string) error {
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
		return fmt.Errorf("unsupported channel type: %s", channel.Type)
	}

	if err != nil {
		return err
	}

	return sender.Send(title, content)
}

// ============ Feishu Sender ============

type FeishuConfig struct {
	WebhookURL string `json:"webhook_url"`
	Secret      string `json:"secret"`
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

func (s *FeishuSender) Send(title, content string) error {
	msg := FeishuMessage{}
	msg.MsgType = "text"
	msg.Content.Text = fmt.Sprintf("**%s**\n%s", title, content)

	body, _ := json.Marshal(msg)
	resp, err := s.client.Post(s.config.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to send feishu notification: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("feishu notification failed with status: %d", resp.StatusCode)
	}
	return nil
}

// ============ DingTalk Sender ============

type DingTalkConfig struct {
	WebhookURL string `json:"webhook_url"`
	Secret      string `json:"secret"`
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

func (s *DingTalkSender) Send(title, content string) error {
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

func (s *WeComSender) Send(title, content string) error {
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
}

type WebhookSender struct {
	config WebhookConfig
	client *http.Client
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
	return &WebhookSender{
		config: wc,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (s *WebhookSender) Send(title, content string) error {
	var body []byte

	// Use template if provided
	if s.config.Template != "" {
		body = []byte(s.config.Template)
	} else {
		// Default JSON format
		payload := map[string]string{
			"title":   title,
			"content": content,
		}
		body, _ = json.Marshal(payload)
	}

	req, err := http.NewRequest(s.config.Method, s.config.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
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
