package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// WorkflowHTTPClient 调用 ai-engine 的 HTTP 工作流引擎 API。
// 工作流能力（/workflows/*）仅通过 ai-engine 的 HTTP 端口暴露，未在 gRPC 协议中提供，
// 因此 agent-service 需要一份独立的 HTTP Client 以桥接「主管 Agent → 子流程」。
type WorkflowHTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewWorkflowHTTPClient(baseURL string) *WorkflowHTTPClient {
	return &WorkflowHTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ExecuteWorkflowRequest 对应 ai-engine POST /workflows/{id}/execute 的请求体。
type ExecuteWorkflowRequest struct {
	InputData map[string]interface{} `json:"input_data"`
}

// ExecuteWorkflowResponse 对应工作流执行接口的返回结构。
type ExecuteWorkflowResponse struct {
	Success    bool                   `json:"success"`
	Output     map[string]interface{} `json:"output"`
	Error      string                 `json:"error"`
	InstanceID string                 `json:"instance_id"`
	Duration   float64                `json:"duration"`
}

// ExecuteWorkflow 在 ai-engine 上执行指定的工作流。
// workflowID 为工作流 ID，tenantID/userID 通过 HTTP 头传给 ai-engine 做租户隔离。
func (c *WorkflowHTTPClient) ExecuteWorkflow(ctx context.Context, workflowID, tenantID, userID string, inputData map[string]interface{}) (*ExecuteWorkflowResponse, error) {
	if inputData == nil {
		inputData = map[string]interface{}{}
	}

	body, err := json.Marshal(ExecuteWorkflowRequest{InputData: inputData})
	if err != nil {
		return nil, fmt.Errorf("marshal workflow request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/workflows/"+workflowID+"/execute", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create workflow request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if tenantID != "" {
		httpReq.Header.Set("X-Tenant-Id", tenantID)
	}
	if userID != "" {
		httpReq.Header.Set("X-User-Id", userID)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("execute workflow request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("execute workflow unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var result ExecuteWorkflowResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decode workflow response: %w", err)
	}
	return &result, nil
}