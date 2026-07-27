import request from './request'

export interface TenantItem {
  id: string
  name: string
  code: string | null
  plan_type: string
  plan_expires_at: string | null
  status: string
  audit_status: string
  audit_remark: string | null
  banned_at: string | null
  banned_reason: string | null
  config: Record<string, any>
  created_at: string
  updated_at: string
}

export interface TenantListResponse {
  items: TenantItem[]
  total: number
  page: number
}

export function getTenants(params: {
  page?: number
  page_size?: number
  search?: string
  status?: string
  audit_status?: string
}): Promise<TenantListResponse> {
  return request.get('/admin/tenants', { params }) as any
}

export function getTenantDetail(id: string): Promise<TenantItem> {
  return request.get(`/admin/tenants/${id}`) as any
}

export function auditTenant(id: string, data: {
  status: string
  remark?: string
}): Promise<TenantItem> {
  return request.put(`/admin/tenants/${id}/audit`, data) as any
}

export function banTenant(id: string, data: {
  banned: boolean
  reason?: string
}): Promise<TenantItem> {
  return request.put(`/admin/tenants/${id}/ban`, data) as any
}
