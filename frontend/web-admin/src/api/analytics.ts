import request from './request'

export interface DashboardStats {
  total_agents: number
  total_conversations: number
  total_messages: number
  total_knowledge_bases: number
  today_conversations?: number
  today_messages?: number
  agent_trend?: number
  conv_trend?: number
  msg_trend?: number
}

export interface TrendDataPoint {
  date: string
  value: number
}

export interface AgentUsage {
  agent_id: string
  agent_name: string
  message_count: number
  conversation_count: number
}

export interface TopIntent {
  intent: string
  count: number
  percentage: number
}

export interface ChannelDistribution {
  channel: string
  count: number
}

export interface RecentConversation {
  id: string
  channel: string
  customer_name: string
  agent_name?: string
  last_message?: string
  status: string
  created_at: string
  message_count: number
}

export function getDashboardStats(): Promise<DashboardStats> {
  return request.get('/chats/analytics/dashboard')
}

export function getConversationTrend(params?: { days?: number }): Promise<TrendDataPoint[]> {
  return request.get('/chats/analytics/trend', { params })
}

export function getAgentUsage(params?: { start_date?: string; end_date?: string }): Promise<AgentUsage[]> {
  return request.get('/chats/analytics/agent-usage', { params })
}

export function getTopIntents(params?: { limit?: number }): Promise<TopIntent[]> {
  return request.get('/ai/analytics/intents', { params })
}

export function getChannelDistribution(): Promise<ChannelDistribution[]> {
  return request.get('/chats/analytics/channel-distribution')
}

export function getRecentConversations(params?: { limit?: number }): Promise<RecentConversation[]> {
  return request.get('/chats/analytics/recent-conversations', { params })
}


export interface QuotaInfo {
  plan_type: string
  plan_name: string
  message_limit: number
  message_used: number
  agent_limit: number
  agent_used: number
  kb_limit: number
  kb_used: number
  plan_expire_at: string | null
  is_free_plan: boolean
}

export function getQuotaInfo(): Promise<QuotaInfo> {
  return request.get('/chats/quota')
}

export function getUnreadCount(): Promise<{ count: number }> {
  return request.get('/chats/unread-count') as any
}
