import request from './request'

export interface Channel {
  id: string
  tenant_id: string
  type: string
  name: string
  config: Record<string, any>
  status: string
  created_at: string
}

export function listChannels(type?: string): Promise<{ items: Channel[]; total: number }> {
  return request.get('/chats/channels', { params: { type } })
}

export function createChannel(type: string, name: string, config: Record<string, any>): Promise<Channel> {
  return request.post('/chats/channels', { type, name, config })
}

export function getChannel(id: string): Promise<Channel> {
  return request.get(`/chats/channels/${id}`)
}

export function saveChannelConfig(id: string, config: Record<string, any>): Promise<Channel> {
  return request.put(`/chats/channels/${id}/config`, { config })
}

export function testChannel(id: string): Promise<{ success: boolean; message: string }> {
  return request.post(`/chats/channels/${id}/test`)
}
