import request from './request'

export interface LoginRequest {
  account: string
  password: string
  captcha_id: string
  captcha_code: string
  invite_code?: string
}

export interface RegisterRequest {
  phone?: string
  email?: string
  password: string
  name?: string
  tenant_name?: string
  captcha_id: string
  captcha_code: string
  invite_code?: string
}

export interface CaptchaResponse {
  captcha_id: string
  captcha_image: string
}

export interface TokenResponse {
  access_token: string
  refresh_token: string
  expires_in: number
}

export interface UserInfo {
  id: string
  phone: string
  email: string
  name: string
  role: string
  tenant_id: string
  avatar_url?: string
}

export interface AuthResponse {
  tokens: TokenResponse
  user: UserInfo
}

export function getCaptcha(): Promise<CaptchaResponse> {
  return request.get('/public/captcha') as any
}

export function login(data: LoginRequest): Promise<AuthResponse> {
  return request.post('/auth/login', data) as any
}

export function register(data: RegisterRequest): Promise<AuthResponse> {
  return request.post('/auth/register', data) as any
}

export function refreshToken(refreshToken: string): Promise<AuthResponse> {
  return request.post('/auth/refresh', { refresh_token: refreshToken }) as any
}

export function changePassword(data: { old_password: string; new_password: string }): Promise<void> {
  return request.post('/users/me/change-password', data) as any
}
