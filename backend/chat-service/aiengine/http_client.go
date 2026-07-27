package aiengine

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 300 * time.Second,
		},
	}
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GenerateReplyRequest struct {
	TenantID       string            `json:"tenant_id"`
	UserID         string            `json:"user_id"`
	Input          string            `json:"input"`
	AgentID        string            `json:"agent_id"`
	ConversationID string            `json:"conversation_id"`
	MessageID      string            `json:"message_id,omitempty"`
	History        []ChatMessage     `json:"history"`
	SystemPrompt   string            `json:"system_prompt"`
	Model          string            `json:"model,omitempty"`
	KnowledgeBaseID string           `json:"knowledge_base_id,omitempty"`
	Context        map[string]string `json:"context,omitempty"`
	TenantConfig   *TenantConfig     `json:"tenant_config,omitempty"`
}

type TenantConfig struct {
	AIEnabled          bool   `json:"ai_enabled"`
	StreamingEnabled   bool   `json:"streaming_enabled"`
	DefaultChatModelID string `json:"default_chat_model_id,omitempty"`
	DefaultProviderID  string `json:"default_provider_id,omitempty"`
	MonthlyTokenLimit  int    `json:"monthly_token_limit"`
	RateLimitPerMinute int    `json:"rate_limit_per_minute"`
}

type GenerateReplyResponse struct {
	Reply      string  `json:"reply"`
	ModelUsed  string  `json:"model_used"`
	TokensUsed int     `json:"tokens_used"`
	LatencyMs  float64 `json:"latency_ms"`
}

func (c *HTTPClient) GenerateReply(ctx context.Context, req GenerateReplyRequest) (*GenerateReplyResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/v1/chat/generate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result GenerateReplyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

type StreamChunk struct {
	Chunk      string `json:"chunk"`
	IsFinal    bool   `json:"is_final"`
	ModelUsed  string `json:"model_used"`
	TokensUsed int    `json:"tokens_used"`
}

func (c *HTTPClient) StreamReply(ctx context.Context, req GenerateReplyRequest) (<-chan StreamChunk, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/v1/chat/stream", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("Cache-Control", "no-cache")
	httpReq.Header.Set("Connection", "keep-alive")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	ch := make(chan StreamChunk, 64)

	go func() {
		defer resp.Body.Close()
		defer close(ch)

		reader := bufio.NewReader(resp.Body)
		var buffer string

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					return
				}
				if line == "" {
					return
				}
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					return
				}

				var chunk StreamChunk
				if err := json.Unmarshal([]byte(data), &chunk); err == nil {
					select {
					case ch <- chunk:
					case <-ctx.Done():
						return
					}
					if chunk.IsFinal {
						return
					}
				}
			} else if strings.HasPrefix(line, "data:") {
				data := strings.TrimPrefix(line, "data:")
				data = strings.TrimSpace(data)
				buffer += data
			}
		}
	}()

	return ch, nil
}
