import request from './request'

export interface AdminAgent {
  id: string
  name: string
  description: string
  agent_type: string
  status: string
  tenant_id: string
  tenant_name?: string
  created_at: string
}

export interface AgentStats {
  total: number
  by_status: Record<string, number>
}

export function getAdminAgents(params?: { page?: number; page_size?: number; tenant_id?: string; status?: string }): Promise<any> {
  return request.get('/admin/agents', { params })
}

export function getAdminAgentStats(): Promise<AgentStats> {
  return request.get('/admin/agents/stats')
}
