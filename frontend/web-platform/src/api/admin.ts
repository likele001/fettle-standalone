import request from './request'

export interface AdminUser {
  id: string
  username: string
  email: string
  role: 'super_admin' | 'admin' | 'operator'
  status: 'active' | 'disabled'
  last_login_at: string | null
  created_at: string
}

export interface AdminListResponse {
  items: AdminUser[]
  total: number
}

export function getAdmins(params?: { page?: number; page_size?: number }): Promise<AdminListResponse> {
  return request.get('/admin/users', { params }) as any
}

export function getAdminDetail(id: string): Promise<AdminUser> {
  return request.get(`/admin/users/${id}`) as any
}

export function createAdmin(data: {
  username: string
  email: string
  password: string
  role: string
}): Promise<AdminUser> {
  return request.post('/admin/users', data) as any
}

export function updateAdmin(id: string, data: Partial<AdminUser>): Promise<AdminUser> {
  return request.put(`/admin/users/${id}`, data) as any
}

export function toggleAdminStatus(id: string, status: 'active' | 'disabled'): Promise<AdminUser> {
  return request.put(`/admin/users/${id}/status`, { status }) as any
}

export function resetAdminPassword(id: string, newPassword: string): Promise<void> {
  return request.post(`/admin/users/${id}/reset-password`, { password: newPassword }) as any
}
