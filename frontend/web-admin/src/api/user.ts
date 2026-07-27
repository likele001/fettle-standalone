import request from './request'

export interface UserInfo {
  id: string
  phone: string
  name: string
  role: string
  tenant_id: string
  avatar_url?: string
  status: string
  created_at: string
}

export interface UpdateUserRequest {
  name?: string
  avatar_url?: string
}

export interface UserListResponse {
  users: UserInfo[]
  total: number
  page: number
  page_size: number
}

// 获取当前用户信息
export function getCurrentUser() {
  return request.get<UserInfo>('/users/me')
}

// 更新当前用户信息
export function updateCurrentUser(data: UpdateUserRequest) {
  return request.put<UserInfo>('/users/me', data)
}

// 获取用户列表
export function getUserList(params: { page?: number; page_size?: number }) {
  return request.get<UserListResponse>('/users', { params })
}

// 创建用户
export function createUser(data: {
  phone: string
  name?: string
  role?: string
}) {
  return request.post<UserInfo>('/users', data)
}

// 获取当前租户信息
export function getCurrentTenant() {
  return request.get<{
    id: string
    name: string
    plan_type: string
    plan_expires_at: string
    status: string
    config: Record<string, any>
    created_at: string
  }>('/tenants/current')
}

// 更新当前租户信息
export function updateCurrentTenant(data: {
  name?: string
  config?: Record<string, any>
}) {
  return request.put('/tenants/current', data)
}
