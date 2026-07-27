import request from './request'

export interface Subscription {
  tenant_id: string
  plan_id: string
  plan_name: string
  status: string
  started_at: string | null
  expires_at: string | null
}

export interface QuotaInfo {
  used: number
  limit: number
  remaining: number
}

export interface Plan {
  id: string
  name: string
  price: number
  currency: string
  description: string
  features: string[]
  quota_limit: number
  max_agents: number
  is_popular?: boolean
  is_active?: boolean
}

export interface Invoice {
  id: string
  tenant_id: string
  plan_name: string
  amount: number
  currency: string
  status: 'pending' | 'paid' | 'failed' | 'cancelled'
  payment_method: string
  paid_at: string | null
  due_date: string
  created_at: string
}

export interface InvoiceListResponse {
  items: Invoice[]
  total: number
}

export interface AdminSubscriptionListResponse {
  items: Subscription[]
  total: number
  page: number
  page_size: number
}

// 获取当前订阅（租户接口）
export function getSubscription(): Promise<Subscription> {
  return request.get('/billing/subscriptions') as any
}

// 获取配额信息（租户接口）
export function getQuota(): Promise<QuotaInfo> {
  return request.get('/billing/quota') as any
}

// 获取可用套餐列表（租户接口）
export function getPlans(): Promise<Plan[]> {
  return request.get('/billing/plans') as any
}

// 获取账单记录（租户接口）
export function getInvoices(params: {
  page?: number
  page_size?: number
  status?: string
}): Promise<InvoiceListResponse> {
  return request.get('/billing/records', { params }) as any
}

// 创建支付订单（租户接口）
export function createPaymentOrder(planId: string): Promise<{
  id: string
  plan_id: string
  plan_name: string
  amount: number
  trade_order_id: string
  status: string
  payment_url: string
}> {
  return request.post('/billing/payment/create', { plan_id: planId }) as any
}

// ========== 管理员接口 ==========

// 管理员获取所有订阅
export function adminGetSubscriptions(params?: { page?: number; page_size?: number }): Promise<AdminSubscriptionListResponse> {
  return request.get('/admin/subscriptions', { params }) as any
}

// 管理员获取所有套餐
export function adminGetPlans(): Promise<{ items: any[] }> {
  return request.get('/admin/plans') as any
}

// 管理员获取所有账单记录
export function adminGetInvoices(params: {
  page?: number
  page_size?: number
  status?: string
}): Promise<InvoiceListResponse> {
  return request.get('/admin/records', { params }) as any
}
