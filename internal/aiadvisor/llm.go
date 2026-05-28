package aiadvisor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LLMConfig holds configuration for the LLM service
type LLMConfig struct {
	BaseURL    string `json:"base_url"`
	Model      string `json:"model"`
	Timeout    int    `json:"timeout"`
	MaxTokens  int    `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// LLMMessage represents a message for the LLM API
type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMRequest represents a request to the LLM API
type LLMRequest struct {
	Model       string       `json:"model"`
	Messages    []LLMMessage `json:"messages"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
	Temperature float64      `json:"temperature,omitempty"`
	Stream      bool         `json:"stream"`
}

// LLMResponse represents a response from the LLM API
type LLMResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      LLMMessage  `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// LLMClient represents an LLM client
type LLMClient struct {
	config     LLMConfig
	httpClient *http.Client
}

// NewLLMClient creates a new LLM client
func NewLLMClient(config LLMConfig) *LLMClient {
	return &LLMClient{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}
}

// GenerateResponse generates a response from the LLM
func (c *LLMClient) GenerateResponse(ctx context.Context, messages []LLMMessage) (string, error) {
	req := LLMRequest{
		Model:       c.config.Model,
		Messages:    messages,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
		Stream:      false,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLM API error: %s - %s", resp.Status, string(body))
	}

	var llmResp LLMResponse
	if err := json.Unmarshal(body, &llmResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(llmResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return llmResp.Choices[0].Message.Content, nil
}

// GenerateStreamResponse generates a streaming response from the LLM
func (c *LLMClient) GenerateStreamResponse(ctx context.Context, messages []LLMMessage, writer io.Writer) error {
	req := LLMRequest{
		Model:       c.config.Model,
		Messages:    messages,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
		Stream:      true,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("LLM API error: %s - %s", resp.Status, string(body))
	}

	_, err = io.Copy(writer, resp.Body)
	return err
}

// HealthCheck checks if the LLM service is healthy
func (c *LLMClient) HealthCheck(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.config.BaseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("LLM service unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("LLM service unhealthy: %s", resp.Status)
	}

	return nil
}
