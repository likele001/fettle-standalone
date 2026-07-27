package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// LongTermMemory 长期记忆
type LongTermMemory struct {
	milvusClient client.Client
	collection   string
}

// NewLongTermMemory 创建长期记忆
func NewLongTermMemory(milvusClient client.Client, collection string) *LongTermMemory {
	return &LongTermMemory{
		milvusClient: milvusClient,
		collection:   collection,
	}
}

// MemoryItem 记忆项
type MemoryItem struct {
	ID           string                 `json:"id"`
	TenantID     string                 `json:"tenant_id"`
	AgentID      string                 `json:"agent_id"`
	UserID       string                 `json:"user_id"`
	Content      string                 `json:"content"`
	MemoryType   string                 `json:"memory_type"` // conversation, fact, preference
	Embedding    []float32              `json:"embedding"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
}

// SaveMemory 保存长期记忆
func (m *LongTermMemory) SaveMemory(ctx context.Context, item *MemoryItem) error {
	// 构建 Milvus 插入数据
	data := []entity.Column{
		entity.NewColumnVarChar("id", []string{item.ID}),
		entity.NewColumnVarChar("tenant_id", []string{item.TenantID}),
		entity.NewColumnVarChar("agent_id", []string{item.AgentID}),
		entity.NewColumnVarChar("user_id", []string{item.UserID}),
		entity.NewColumnVarChar("content", []string{item.Content}),
		entity.NewColumnVarChar("memory_type", []string{item.MemoryType}),
		entity.NewColumnFloatVector("embedding", 1536, [][]float32{item.Embedding}),
	}

	metadataJSON, _ := json.Marshal(item.Metadata)
	data = append(data, entity.NewColumnVarChar("metadata", []string{string(metadataJSON)}))

	_, err := m.milvusClient.Insert(ctx, m.collection, "", data)
	return err
}

// Retrieve 检索相关记忆
func (m *LongTermMemory) Retrieve(ctx context.Context, tenantID string, queryEmbedding []float32, topK int) ([]*MemoryItem, error) {
	// 构建搜索向量
	vectors := []entity.Vector{
		entity.FloatVector(queryEmbedding),
	}

	// 搜索参数
	sp := entity.NewIndexFlatSearchParam()
	
	// 执行搜索
	searchResult, err := m.milvusClient.Search(
		ctx,
		m.collection,
		[]string{},
		fmt.Sprintf("tenant_id == \"%s\"", tenantID),
		[]string{"id", "tenant_id", "agent_id", "user_id", "content", "memory_type", "metadata"},
		vectors,
		"embedding",
		entity.L2,
		topK,
		sp,
	)
	if err != nil {
		return nil, err
	}

	// 解析结果
	var memories []*MemoryItem
	for _, result := range searchResult {
		for i := 0; i < result.ResultCount; i++ {
			id, _ := result.Fields.GetByName("id").Get(i).(string)
			tenantID, _ := result.Fields.GetByName("tenant_id").Get(i).(string)
			agentID, _ := result.Fields.GetByName("agent_id").Get(i).(string)
			userID, _ := result.Fields.GetByName("user_id").Get(i).(string)
			content, _ := result.Fields.GetByName("content").Get(i).(string)
			memoryType, _ := result.Fields.GetByName("memory_type").Get(i).(string)
			metadataJSON, _ := result.Fields.GetByName("metadata").Get(i).(string)

			var metadata map[string]interface{}
			json.Unmarshal([]byte(metadataJSON), &metadata)

			memories = append(memories, &MemoryItem{
				ID:         id,
				TenantID:   tenantID,
				AgentID:    agentID,
				UserID:     userID,
				Content:    content,
				MemoryType: memoryType,
				Metadata:   metadata,
			})
		}
	}

	return memories, nil
}

// SummarizeConversation 生成会话摘要
func (m *LongTermMemory) SummarizeConversation(ctx context.Context, messages []map[string]string) (string, error) {
	// TODO: 调用 LLM 生成会话摘要
	// 这里简单返回最后几条消息的拼接
	if len(messages) == 0 {
		return "", nil
	}

	summary := ""
	for i := len(messages) - 3; i < len(messages); i++ {
		if i >= 0 {
			role := messages[i]["role"]
			content := messages[i]["content"]
			summary += fmt.Sprintf("%s: %s\n", role, content)
		}
	}

	return summary, nil
}

// DeleteMemory 删除记忆
func (m *LongTermMemory) DeleteMemory(ctx context.Context, memoryID string) error {
	// Milvus 删除操作
	expr := fmt.Sprintf("id == \"%s\"", memoryID)
	return m.milvusClient.Delete(ctx, m.collection, "", expr)
}
