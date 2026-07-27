import { get, post, put, del } from '@/utils/request'

export interface Agent {
  id: string
  name: string
  description: string
  avatar: string
  model_id: string
  system_prompt: string
  industry: string
  status: 'active' | 'inactive'
  created_at: string
  updated_at: string
}

export interface AgentListParams {
  page?: number
  page_size?: number
  status?: string
  keyword?: string
  industry?: string
}

export interface AgentListResponse {
  items: Agent[]
  total: number
  page: number
  page_size: number
}

// 获取智能体列表
export function getAgents(params?: AgentListParams) {
  return get<AgentListResponse>('/agents', params)
}

// 获取智能体详情
export function getAgentDetail(id: string) {
  return get<Agent>(`/agents/${id}`)
}

// 创建智能体
export function createAgent(data: Partial<Agent>) {
  return post<Agent>('/agents', data)
}

// 更新智能体
export function updateAgent(id: string, data: Partial<Agent>) {
  return put<Agent>(`/agents/${id}`, data)
}

// 删除智能体
export function deleteAgent(id: string) {
  return del(`/agents/${id}`)
}
