import request from './request'

export interface PlatformOverview {
  total_tenants: number
  total_agents: number
  total_conversations: number
  total_messages: number
  active_channels: number
}

export function getPlatformOverview(): Promise<PlatformOverview> {
  return request.get('/admin/analytics/overview')
}
