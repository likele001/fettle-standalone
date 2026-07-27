import request from './request'

export interface Plan {
  id: string
  name: string
  type: 'free' | 'pro' | 'enterprise'
  description: string
  price_monthly: number
  price_yearly: number
  max_agents: number
  max_knowledge_bases: number
  max_documents_per_kb: number
  max_messages_per_month: number
  max_concurrent_sessions: number
  features: string[]
  allowed_industries: string[]
  allowed_features: string[]
  is_active: boolean
  is_default: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

export interface PlanListResponse {
  items: Plan[]
}

// 获取所有套餐
export function getPlans(): Promise<PlanListResponse> {
  return request.get('/admin/plans') as any
}

// 创建套餐
export function createPlan(data: Partial<Plan>): Promise<Plan> {
  return request.post('/admin/plans', data) as any
}

// 更新套餐
export function updatePlan(id: string, data: Partial<Plan>): Promise<Plan> {
  return request.put(`/admin/plans/${id}`, data) as any
}

// 删除套餐
export function deletePlan(id: string): Promise<void> {
  return request.delete(`/admin/plans/${id}`) as any
}

// 切换套餐状态
export function togglePlanStatus(id: string, status: 'active' | 'inactive'): Promise<void> {
  return request.put(`/admin/plans/${id}/status`, { status }) as any
}
