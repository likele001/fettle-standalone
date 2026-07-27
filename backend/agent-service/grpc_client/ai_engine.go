package grpc_client

import (
	"context"
	"fmt"
	"time"

	"ai-platform/agent-service/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type AIEngineClient struct {
	conn    *grpc.ClientConn
	client  proto.AIEngineClient
	addr    string
	timeout time.Duration
}

func NewAIEngineClient(addr string, timeout time.Duration) *AIEngineClient {
	return &AIEngineClient{
		addr:    addr,
		timeout: timeout,
	}
}

func (c *AIEngineClient) Connect() error {
	conn, err := grpc.Dial(c.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to AI Engine: %w", err)
	}
	c.conn = conn
	c.client = proto.NewAIEngineClient(conn)
	return nil
}

func (c *AIEngineClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *AIEngineClient) ctxWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if c.timeout > 0 {
		return context.WithTimeout(ctx, c.timeout)
	}
	return ctx, func() {}
}

func (c *AIEngineClient) RecognizeIntent(ctx context.Context, tenantID, userID, input, agentID string) (string, float64, error) {
	ctx, cancel := c.ctxWithTimeout(ctx)
	defer cancel()

	ctx = metadata.AppendToOutgoingContext(ctx, "tenant_id", tenantID)

	resp, err := c.client.RecognizeIntent(ctx, &proto.IntentRequest{
		TenantId: tenantID,
		UserId:   userID,
		Input:    input,
		AgentId:  agentID,
	})
	if err != nil {
		return "unknown", 0.0, fmt.Errorf("recognize intent failed: %w", err)
	}

	return resp.Intent, float64(resp.Confidence), nil
}

func (c *AIEngineClient) GenerateReply(ctx context.Context, tenantID, userID, input, agentID, conversationID string, history []map[string]string) (string, error) {
	ctx, cancel := c.ctxWithTimeout(ctx)
	defer cancel()

	ctx = metadata.AppendToOutgoingContext(ctx, "tenant_id", tenantID)

	var protoHistory []*proto.Message
	for _, msg := range history {
		role, _ := msg["role"]
		content, _ := msg["content"]
		protoHistory = append(protoHistory, &proto.Message{
			Role:    role,
			Content: content,
		})
	}

	resp, err := c.client.GenerateReply(ctx, &proto.ReplyRequest{
		TenantId:       tenantID,
		UserId:         userID,
		Input:          input,
		AgentId:        agentID,
		ConversationId: conversationID,
		History:        protoHistory,
	})
	if err != nil {
		return "", fmt.Errorf("generate reply failed: %w", err)
	}

	return resp.Reply, nil
}

func (c *AIEngineClient) PlanTask(ctx context.Context, tenantID, userID, goal, agentID string, tools []string) ([]map[string]interface{}, error) {
	ctx, cancel := c.ctxWithTimeout(ctx)
	defer cancel()

	ctx = metadata.AppendToOutgoingContext(ctx, "tenant_id", tenantID)

	resp, err := c.client.PlanTask(ctx, &proto.PlanRequest{
		TenantId:       tenantID,
		UserId:         userID,
		Goal:           goal,
		AgentId:        agentID,
		AvailableTools: tools,
	})
	if err != nil {
		return nil, fmt.Errorf("plan task failed: %w", err)
	}

	var steps []map[string]interface{}
	for _, step := range resp.Steps {
		params := make(map[string]string)
		if step.Parameters != nil {
			params = step.Parameters
		}
		steps = append(steps, map[string]interface{}{
			"step_id":     step.StepId,
			"action":      step.Action,
			"tool":        step.Tool,
			"parameters":  params,
			"description": step.Description,
		})
	}
	return steps, nil
}
