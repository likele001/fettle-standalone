import request from './request'

// 厂商管理
export function getProviders(params?: { status?: string }) {
  return request.get('/admin/ai/providers', { params }) as any
}

export function getProvider(id: string) {
  return request.get(`/admin/ai/providers/${id}`)
}

export function createProvider(data: any) {
  return request.post('/admin/ai/providers', data)
}

export function updateProvider(id: string, data: any) {
  return request.put(`/admin/ai/providers/${id}`, data)
}

export function deleteProvider(id: string) {
  return request.delete(`/admin/ai/providers/${id}`)
}

// 模型管理
export function getModels(params?: { provider_id?: string; status?: string }) {
  return request.get('/admin/ai/models', { params }) as any
}

export function getModel(id: string) {
  return request.get(`/admin/ai/models/${id}`)
}

export function createModel(data: any) {
  return request.post('/admin/ai/models', data)
}

export function updateModel(id: string, data: any) {
  return request.put(`/admin/ai/models/${id}`, data)
}

export function deleteModel(id: string) {
  return request.delete(`/admin/ai/models/${id}`)
}

// 预置厂商和模型
export function seedProviders() {
  return request.post('/admin/ai/providers/seed')
}
