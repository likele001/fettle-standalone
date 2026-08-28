package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// AIBillingClient calls the billing-service internal API to deduct AI costs
// for tenants using platform-managed AI mode (no custom API key).
type AIBillingClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewAIBillingClient creates a new billing client.
// baseURL is the internal billing-service address, e.g. "http://localhost:9600".
func NewAIBillingClient(baseURL string) *AIBillingClient {
	return &AIBillingClient{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		baseURL:    baseURL,
	}
}

type deductAICostRequest struct {
	TenantID     string `json:"tenant_id"`
	ModelID      string `json:"model_id"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
}

type deductAICostResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// DeductAICost calls the billing-service internal endpoint to deduct AI usage cost.
// Returns nil on success, error on failure (e.g. insufficient balance).
func (c *AIBillingClient) DeductAICost(tenantID, modelID string, inputTokens, outputTokens int) error {
	reqBody := deductAICostRequest{
		TenantID:     tenantID,
		ModelID:      modelID,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal deduct request: %w", err)
	}

	url := c.baseURL + "/billing/internal/deduct-ai-cost"
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("call billing-service: %w", err)
	}
	defer resp.Body.Close()

	var result deductAICostResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode deduct response: %w", err)
	}

	if resp.StatusCode == http.StatusPaymentRequired {
		return fmt.Errorf("insufficient balance: %s", result.Message)
	}

	if !result.Success {
		return fmt.Errorf("deduct failed: %s", result.Message)
	}

	log.Printf("AI cost deducted for tenant %s: model=%s input=%d output=%d", tenantID, modelID, inputTokens, outputTokens)
	return nil
}
