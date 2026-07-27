import request from './request'

export interface Conversation {
  id: string
  tenant_id: string
  channel: string
  customer_id: string
  customer_name: string
  agent_id: string
  status: string
  message_count: number
  created_at: string
  updated_at: string
}

export interface Message {
  id: string
  conversation_id: string
  role: string
  content: string
  content_type: string
  created_at: string
}

export function listConversations(params?: {
  page?: number
  page_size?: number
  status?: string
  agent_id?: string
  channel?: string
}): Promise<{ items: Conversation[]; total: number }> {
  return request.get('/chats/conversations', { params })
}

export function getConversation(id: string): Promise<Conversation> {
  return request.get(`/chats/conversations/${id}`)
}

export function closeConversation(id: string): Promise<void> {
  return request.post(`/chats/conversations/${id}/close`)
}

export function getMessages(conversationId: string, params?: { page?: number; page_size?: number }): Promise<{ items: Message[]; total: number }> {
  return request.get(`/chats/conversations/${conversationId}/messages`, { params })
}

export function sendMessage(conversationId: string, data: { content: string; content_type?: string }): Promise<Message> {
  return request.post(`/chats/conversations/${conversationId}/messages`, data)
}

export function searchMessages(conversationId: string, query: string): Promise<{ items: Message[]; total: number }> {
  return request.get(`/chats/conversations/${conversationId}/messages/search`, { params: { q: query } })
}

export function takeOverConversation(conversationId: string): Promise<void> {
  return request.post(`/chats/conversations/${conversationId}/takeover`)
}
