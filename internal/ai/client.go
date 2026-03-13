package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/game-ops/ai-alert-system/internal/config"
)

type Client struct {
	httpClient *http.Client
	Config    *config.AIConfig
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string      `json:"model"`
	Messages    []Message   `json:"messages"`
	MaxTokens   int         `json:"max_tokens,omitempty"`
	Temperature float64     `json:"temperature,omitempty"`
	ReasoningSplit bool     `json:"reasoning_split,omitempty"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func NewClient(cfg *config.AIConfig) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		Config: cfg,
	}
}

// Chat sends a chat request to the AI API (OpenAI compatible)
func (c *Client) Chat(systemPrompt, userMessage string) (string, error) {
	if c.Config.APIKey == "" {
		return "", fmt.Errorf("AI API key not configured")
	}

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	return c.chat(messages)
}

// ChatWithContext sends a chat request with conversation history
func (c *Client) ChatWithContext(messages []Message, systemPrompt string) (string, error) {
	if c.Config.APIKey == "" {
		return "", fmt.Errorf("AI API key not configured")
	}

	// Prepend system message
	fullMessages := []Message{
		{Role: "system", Content: systemPrompt},
	}
	fullMessages = append(fullMessages, messages...)

	return c.chat(fullMessages)
}

func (c *Client) chat(messages []Message) (string, error) {
	reqBody := ChatRequest{
		Model:          c.Config.Model,
		Messages:       messages,
		MaxTokens:      2000,
		Temperature:    0.7,
		ReasoningSplit: false, // 关闭思考过程
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build API URL - support both OpenAI and MiniMax
	apiBase := c.Config.APIBase
	if apiBase == "" {
		apiBase = "https://api.minimaxi.com/v1"
	}

	// Ensure base URL ends with /v1
	if !hasSuffix(apiBase, "/v1") {
		apiBase = apiBase + "/v1"
	}

	apiPath := "/chat/completions"
	req, err := http.NewRequest("POST", apiBase+apiPath, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Config.APIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return "", fmt.Errorf("API request failed with status %d", resp.StatusCode)
		}
		return "", fmt.Errorf("API error: %s", errResp.Error.Message)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// AnalyzeAlert analyzes an alert and returns suggestions
func (c *Client) AnalyzeAlert(alertName, message, labels string) (string, string, []string, error) {
	prompt := fmt.Sprintf(`你是一个游戏运维告警分析助手。请分析以下告警并提供：
1. 简要总结 (ai_summary)
2. 可能根因 (ai_root_cause)
3. 处置建议列表 (ai_suggestions)

告警名称: %s
告警消息: %s
Labels: %s

请以 JSON 格式返回：
{"summary": "...", "root_cause": "...", "suggestions": ["...", "..."]}`, alertName, message, labels)

	response, err := c.Chat(prompt, "")
	if err != nil {
		return "", "", nil, err
	}

	return response, "", nil, nil
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
