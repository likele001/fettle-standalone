import { get, post, put, del } from '@/utils/request'

export interface KnowledgeBase {
  id: string
  name: string
  description: string
  document_count: number
  status: 'active' | 'processing' | 'error'
  created_at: string
  updated_at: string
}

export interface Document {
  id: string
  knowledge_base_id: string
  name: string
  type: string
  size: number
  status: 'pending' | 'processing' | 'completed' | 'error'
  created_at: string
}

export interface KnowledgeBaseListParams {
  page?: number
  page_size?: number
  status?: string
  keyword?: string
}

export interface KnowledgeBaseListResponse {
  items: KnowledgeBase[]
  total: number
  page: number
  page_size: number
}

export interface DocumentListResponse {
  items: Document[]
  total: number
}

// 获取知识库列表
export function getKnowledgeBases(params?: KnowledgeBaseListParams) {
  return get<KnowledgeBaseListResponse>('/knowledge-bases', params)
}

// 获取知识库详情
export function getKnowledgeBaseDetail(id: string) {
  return get<KnowledgeBase>(`/knowledge-bases/${id}`)
}

// 创建知识库
export function createKnowledgeBase(data: Partial<KnowledgeBase>) {
  return post<KnowledgeBase>('/knowledge-bases', data)
}

// 更新知识库
export function updateKnowledgeBase(id: string, data: Partial<KnowledgeBase>) {
  return put<KnowledgeBase>(`/knowledge-bases/${id}`, data)
}

// 删除知识库
export function deleteKnowledgeBase(id: string) {
  return del(`/knowledge-bases/${id}`)
}

// 获取文档列表
export function getDocuments(knowledge_base_id: string, page = 1, page_size = 20) {
  return get<DocumentListResponse>(`/knowledge-bases/${knowledge_base_id}/documents`, {
    page,
    page_size
  })
}

// 上传文档
export function uploadDocument(knowledge_base_id: string, file_path: string) {
  return post<Document>(`/knowledge-bases/${knowledge_base_id}/documents`, {
    file_path
  })
}

// 删除文档
export function deleteDocument(knowledge_base_id: string, document_id: string) {
  return del(`/knowledge-bases/${knowledge_base_id}/documents/${document_id}`)
}
