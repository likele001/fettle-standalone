import request from './request'

export function getAdminChannels(params?: { page?: number; page_size?: number; tenant_id?: string }): Promise<any> {
  return request.get('/admin/channels', { params })
}

export function getAdminChannelStats(): Promise<any> {
  return request.get('/admin/channels/stats')
}
