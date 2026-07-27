import request from './request'

// ===== 租户级 API =====

export interface TenantAIConfig {
  default_provider_id?: string
  default_chat_model_id?: string
  default_embedding_model_id?: string
  config?: {
    temperature?: number
    max_tokens?: number
  }
}

export function getTenantAIConfig(): Promise<TenantAIConfig> {
  return request.get('/tenant/ai/config') as any
}

// 更新租户 AI 配置
export function updateTenantAIConfig(data: any) {
  return request.put('/tenant/ai/config', data)
}

// 获取可用厂商列表（租户视角）
export function getProviders() {
  return request.get('/tenant/ai/providers')
}

// 获取厂商模型列表（租户视角）
export function getModels(providerId?: string) {
  return request.get('/tenant/ai/models', { params: { provider_id: providerId } })
}

// 获取租户 API Keys
export function getAPIKeys() {
  return request.get('/tenant/ai/api-keys')
}

// 创建 API Key
export function createAPIKey(data: { provider_id: string; key_name: string; api_key: string; custom_base_url?: string }) {
  return request.post('/tenant/ai/api-keys', data)
}

// 删除 API Key
export function deleteAPIKey(id: string) {
  return request.delete(`/tenant/ai/api-keys/${id}`)
}

export interface APIKeyTestResult {
  valid?: boolean
  success?: boolean
  error?: string
}

export function testAPIKey(id: string): Promise<APIKeyTestResult> {
  return request.post(`/tenant/ai/api-keys/${id}/test`) as any
}

// ===== 管理员级 API（模型 CRUD）=====

// 获取所有模型（管理员视角，含全部厂商）
export function adminGetModels(params?: { provider_id?: string; model_type?: string }) {
  return request.get('/admin/ai/models', { params })
}

// 获取单个模型详情
export function adminGetModel(id: string) {
  return request.get(`/admin/ai/models/${id}`)
}

// 创建模型
export function adminCreateModel(data: any) {
  return request.post('/admin/ai/models', data)
}

// 更新模型
export function adminUpdateModel(id: string, data: any) {
  return request.put(`/admin/ai/models/${id}`, data)
}

// 删除模型
export function adminDeleteModel(id: string) {
  return request.delete(`/admin/ai/models/${id}`)
}

// 获取所有厂商（管理员视角）
export function adminGetProviders() {
  return request.get('/admin/ai/providers')
}
