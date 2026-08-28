package aiengine

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

// WorkflowClient 调用 ai-engine 的工作流引擎 API。
// 通过 HTTP 触发智能体绑定的工作流子流程，实现「主管 Agent → 子流程」多智能体协同。
type WorkflowClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewWorkflowClient(baseURL string) *WorkflowClient {
	return &WorkflowClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type ExecuteWorkflowRequest struct {
	InputData map[string]interface{} `json:"input_data"`
}

type ExecuteWorkflowResponse struct {
	Success    bool                   `json:"success"`
	Output     map[string]interface{} `json:"output"`
	Error      string                 `json:"error"`
	InstanceID string                 `json:"instance_id"`
	Duration   float64                `json:"duration"`
}

// ExecuteWorkflow 执行指定的工作流。
func (c *WorkflowClient) ExecuteWorkflow(ctx context.Context, workflowID, tenantID, userID string, inputData map[string]interface{}) (*ExecuteWorkflowResponse, error) {
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