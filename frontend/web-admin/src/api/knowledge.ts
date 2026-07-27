import request from './request'

export interface KnowledgeBase {
  id: string
  tenant_id: string
  name: string
  description: string
  status: string
  doc_count: number
  chunk_count: number
  total_size: number
  config: Record<string, any>
  created_at: string
  updated_at: string
}

export interface KnowledgeDocument {
  id: string
  knowledge_base_id: string
  tenant_id: string
  file_name: string
  file_type: string
  file_size: number
  file_url: string
  status: string
  chunk_count: number
  error_message: string
  processed_at: string | null
  created_at: string
  updated_at: string
}

export interface ListResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export function getKnowledgeBases(page = 1, pageSize = 20): Promise<ListResponse<KnowledgeBase>> {
  return request.get('/knowledge', {
    params: { page, page_size: pageSize }
  })
}

// 获取知识库详情
export function getKnowledgeBase(id: string) {
  return request.get<KnowledgeBase>(`/knowledge/${id}`)
}

// 创建知识库
export function createKnowledgeBase(data: { name: string; description?: string }) {
  return request.post<KnowledgeBase>('/knowledge', data)
}

// 更新知识库
export function updateKnowledgeBase(id: string, data: { name?: string; description?: string }) {
  return request.put<KnowledgeBase>(`/knowledge/${id}`, data)
}

// 删除知识库
export function deleteKnowledgeBase(id: string) {
  return request.delete(`/knowledge/${id}`)
}

export function getDocuments(kbId: string, page = 1, pageSize = 20): Promise<ListResponse<KnowledgeDocument>> {
  return request.get(`/knowledge/${kbId}/documents`, {
    params: { page, page_size: pageSize }
  })
}

// 上传文档
export function uploadDocument(kbId: string, file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return request.post<KnowledgeDocument>(`/knowledge/${kbId}/documents`, formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}

// 删除文档
export function deleteDocument(kbId: string, docId: string) {
  return request.delete(`/knowledge/${kbId}/documents/${docId}`)
}
