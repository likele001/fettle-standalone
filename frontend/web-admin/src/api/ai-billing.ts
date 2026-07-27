import request from './request'

// ===== 类型定义 =====

export interface BalanceInfo {
  balance: number
  total_recharged: number
  total_consumed: number
  active_packages: ResourcePackage[]
}

export interface ResourcePackage {
  id: string
  name: string
  remaining_tokens: number
  total_tokens: number
  expires_at: string
  status: string
}

export interface AvailablePackage {
  id: string
  name: string
  tokens: number
  price: number
  description?: string
}

export interface UsageDetail {
  id: string
  date: string
  model_name: string
  input_tokens: number
  output_tokens: number
  total_tokens: number
  cost: number
  billing_mode: string
}

export interface UsageDetailResponse {
  items: UsageDetail[]
  total: number
  page: number
  page_size: number
}

export interface ModelPricing {
  id: string
  model_name: string
  input_price: number
  output_price: number
}

// ===== API 调用 =====

// 获取余额与活跃资源包
export function getBalance(): Promise<BalanceInfo> {
  return request.get('/billing/balance') as any
}

// 获取可购买的资源包列表
export function getPackages(): Promise<AvailablePackage[]> {
  return request.get('/billing/packages') as any
}

// 购买资源包
export function purchasePackage(packageId: string) {
  return request.post('/billing/packages/purchase', { package_id: packageId })
}

// 获取用量明细
export function getUsageDetails(params: {
  page?: number
  page_size?: number
  start_date?: string
  end_date?: string
}): Promise<UsageDetailResponse> {
  return request.get('/billing/usage-details', { params }) as any
}

// 获取平台模型定价
export function getPlatformPricing(): Promise<ModelPricing[]> {
  return request.get('/billing/platform-pricing') as any
}
