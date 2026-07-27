import request from './request'

export interface Agent {
  id: string
  tenant_id: string
  name: string
  description: string
  avatar_url: string
  status: string
  agent_type: string
  personality_config: Record<string, any>
  workflow_config: Record<string, any>
  welcome_message: string
  work_hours: Record<string, any>
  model_id: string
  knowledge_base_ids: string[]
  max_concurrent: number
  current_sessions: number
  total_conversations: number
  total_messages: number
  created_at: string
  updated_at: string
}

export interface CreateAgentRequest {
  name: string
  description?: string
  avatar_url?: string
  agent_type?: string
  personality_config?: Record<string, any>
  workflow_config?: Record<string, any>
  welcome_message?: string
  work_hours?: Record<string, any>
  model_id?: string
  knowledge_base_ids?: string[]
  max_concurrent?: number
}

export interface UpdateAgentRequest {
  name?: string
  description?: string
  avatar_url?: string
  personality_config?: Record<string, any>
  workflow_config?: Record<string, any>
  welcome_message?: string
  work_hours?: Record<string, any>
  model_id?: string
  knowledge_base_ids?: string[]
  max_concurrent?: number
}

export interface ListResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export function getAgents(page = 1, pageSize = 20, params?: Record<string, any>): Promise<ListResponse<Agent>> {
  return request.get('/agents', {
    params: { page, page_size: pageSize, ...params }
  })
}

// 获取智能体详情
export function getAgent(id: string) {
  return request.get<Agent>(`/agents/${id}`)
}

// 创建智能体
export function createAgent(data: CreateAgentRequest) {
  return request.post<Agent>('/agents', data)
}

// 更新智能体
export function updateAgent(id: string, data: UpdateAgentRequest) {
  return request.put<Agent>(`/agents/${id}`, data)
}

// 删除智能体
export function deleteAgent(id: string) {
  return request.delete(`/agents/${id}`)
}

export interface TestChatResponse {
  reply: string
}

export function testChat(agentId: string, message: string): Promise<TestChatResponse> {
  return request.post('/agents/test', { agent_id: agentId, message })
}
