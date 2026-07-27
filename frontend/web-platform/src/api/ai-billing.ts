import request from './request'

// ========== 模型定价 ==========

export interface ModelPricing {
  model_id: string
  model_name: string
  provider_name: string
  model_type: string
  input_price_per_1k: number
  output_price_per_1k: number
  status: string
}

export function getPricingList(): Promise<ModelPricing[]> {
  return request.get('/admin/ai/pricing') as any
}

export function updatePricing(modelId: string, data: {
  input_price_per_1k?: number
  output_price_per_1k?: number
  status?: string
}): Promise<any> {
  return request.put(`/admin/ai/pricing/${modelId}`, data) as any
}

// ========== 租户余额 ==========

export interface TenantBalance {
  tenant_id: string
  tenant_name: string
  balance: number
  total_recharged: number
  total_consumed: number
  updated_at: string
}

export function getBalancesList(params?: { keyword?: string }): Promise<TenantBalance[]> {
  return request.get('/admin/ai/balances', { params }) as any
}

export function rechargeTenant(data: {
  tenant_id: string
  amount: number
  remark?: string
}): Promise<any> {
  return request.post('/admin/ai/recharge', data) as any
}

// ========== 资源包 ==========

export interface ResourcePackage {
  id: string
  name: string
  token_amount: number
  price: number
  description: string
  status: string
  created_at: string
}

export function getPackagesList(): Promise<ResourcePackage[]> {
  return request.get('/admin/ai/packages') as any
}

export function createPackage(data: {
  name: string
  token_amount: number
  price: number
  description?: string
}): Promise<any> {
  return request.post('/admin/ai/packages', data) as any
}

// ========== 用量统计 ==========

export interface UsageStats {
  total_tokens: number
  total_cost: number
  total_requests: number
  period: string
}

export function getUsageStats(params?: {
  start_date?: string
  end_date?: string
}): Promise<UsageStats> {
  return request.get('/admin/ai/usage', { params }) as any
}
