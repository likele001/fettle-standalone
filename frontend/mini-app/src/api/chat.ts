import { get, post } from '@/utils/request'

export interface Message {
  id: string
  conversation_id: string
  role: 'user' | 'assistant' | 'system'
  content: string
  created_at: string
}

export interface Conversation {
  id: string
  agent_id: string
  agent_name: string
  title: string
  status: 'active' | 'closed'
  last_message: string
  created_at: string
  updated_at: string
}

export interface SendMessageParams {
  conversation_id: string
  content: string
}

export interface ConversationListParams {
  page?: number
  page_size?: number
  agent_id?: string
  status?: string
}

export interface ConversationListResponse {
  items: Conversation[]
  total: number
  page: number
  page_size: number
}

export interface MessageListResponse {
  items: Message[]
  total: number
}

// 获取会话列表
export function getConversations(params?: ConversationListParams) {
  return get<ConversationListResponse>('/chats/conversations', params)
}

// 获取会话详情
export function getConversationDetail(id: string) {
  return get<Conversation>(`/conversations/${id}`)
}

// 创建会话
export function createConversation(agent_id: string) {
  return post<Conversation>('/conversations', { agent_id })
}

// 发送消息
export function sendMessage(params: SendMessageParams) {
  return post<Message>('/conversations/messages', params)
}

// 获取消息列表
export function getMessages(conversation_id: string, page = 1, page_size = 50) {
  return get<MessageListResponse>(`/conversations/${conversation_id}/messages`, {
    page,
    page_size
  })
}

// 关闭会话
export function closeConversation(id: string) {
  return post(`/conversations/${id}/close`)
}
