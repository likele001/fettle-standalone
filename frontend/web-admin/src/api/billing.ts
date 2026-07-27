import request from './request'

// 套餐信息
export interface Plan {
  id: string
  name: string
  description: string
  price: number
  period: string // month, year
  max_agents: number
  max_messages: number
  features: string[]
  allowed_industries: string[]
  allowed_features: string[]
  is_active: boolean
  created_at: string
}

// 订阅信息
export interface Subscription {
  id: string
  tenant_id: string
  plan_id: string
  plan?: Plan
  status: string // active, expired, cancelled
  start_date: string
  end_date: string
  auto_renew: boolean
  messages_used: number
  created_at: string
}

// 计费记录
export interface BillingRecord {
  id: string
  tenant_id: string
  subscription_id: string
  type: string // message, agent, storage
  amount: number
  cost: number
  record_date: string
  created_at: string
}

// 配额信息
export interface Quota {
  used: number
  limit: number
  remaining: number
}

// 月度使用量
export interface MonthlyUsage {
  usage: number
  year: number
  month: number
}

// 支付订单
export interface PaymentOrder {
  id: string
  plan_id: string
  plan_name: string
  amount: number
  trade_order_id: string
  status: string
  payment_url: string
  transaction_id: string
  paid_at: string | null
  created_at: string
}

export interface PlanListResponse {
  items: Plan[]
}

export function getPlans(): Promise<PlanListResponse> {
  return request.get('/billing/plans') as any
}

// 获取套餐详情
export function getPlan(id: string) {
  return request.get(`/billing/plans/${id}`)
}

// 获取当前订阅
export function getSubscription() {
  return request.get('/billing/subscriptions')
}

// 订阅套餐（免费套餐直接激活）
export function subscribe(planId: string) {
  return request.post('/billing/subscriptions', { plan_id: planId })
}

// 创建支付订单（付费套餐，返回支付链接）
export function createPayment(planId: string) {
  return request.post('/billing/payment/create', { plan_id: planId })
}

// 获取配额
export function getQuota() {
  return request.get('/billing/quota')
}

// 获取月度使用量
export function getMonthlyUsage(year?: number, month?: number) {
  return request.get('/billing/usage', {
    params: { year, month }
  })
}

// 获取账单列表
export function getBills(params?: { status?: string; page?: number; page_size?: number }) {
  return request.get('/billing/records', { params })
}

// 获取账单详情
export function getBillDetail(id: string) {
  return request.get(`/billing/bills/${id}`)
}

// 获取支付订单列表
export function getPaymentOrders(params?: { page?: number; page_size?: number }) {
  return request.get('/billing/payment/orders', { params })
}
