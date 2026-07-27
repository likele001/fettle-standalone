import request from './request'

interface LoginRequest {
  account: string
  password: string
  captcha_id: string
  captcha_code: string
}

interface CaptchaResponse {
  captcha_id: string
  captcha_image: string
}

interface TokenResponse {
  access_token: string
  refresh_token: string
  expires_in: number
}

interface UserInfo {
  id: string
  username: string
  phone: string
  name: string
  role: string
  email?: string
  avatar_url?: string
}

interface AuthResponse {
  tokens: TokenResponse
  user: UserInfo
}

export function getCaptcha(): Promise<CaptchaResponse> {
  return request.get('/public/captcha') as any
}

export function login(data: LoginRequest): Promise<AuthResponse> {
  return request.post('/admin/auth/login', data) as any
}

export function refreshToken(refreshToken: string): Promise<AuthResponse> {
  return request.post('/admin/auth/refresh', { refresh_token: refreshToken }) as any
}

export function logout(): Promise<void> {
  return request.post('/admin/auth/logout') as any
}

export function changePassword(data: { old_password: string; new_password: string }): Promise<void> {
  return request.post('/admin/auth/change-password', data) as any
}
