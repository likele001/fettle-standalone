
export interface MiniappLoginResult {
  token?: string
  need_bind?: boolean
  openid?: string
  user?: {
    id: string
    name: string
    phone: string
    email: string
    avatar_url: string
    role: string
    tenant_id: string
  }
}

export interface BindOpenidParams {
  username: string
  password: string
  openid: string
  tenant_code: string
}

export interface MiniAppConfig {
  app_id: string
}

// 获取小程序配置（微信 AppID 等）
export function fetchMiniConfig() {
  return get<MiniAppConfig>('/public/miniapp/config')
}

// 小程序微信登录
export function miniappLogin(code: string, tenantCode: string) {
  return get<MiniappLoginResult>(
    `/miniapp/auth/login?code=${encodeURIComponent(code)}&tenant_code=${encodeURIComponent(tenantCode)}`
  )
}

// 绑定已有账号到微信 openid
export function bindOpenid(params: BindOpenidParams) {
  return post<MiniappLoginResult>('/miniapp/auth/bind-openid', params)
}

// 获取当前用户信息
export function getMiniappUserInfo() {
  return get<{
    id: string
    name: string
    phone: string
    email: string
    avatar_url: string
    role: string
    tenant_id: string
  }>('/miniapp/user/info')
}

// 刷新 Token
export function refreshMiniappToken(refreshToken: string) {
  return post<{ access_token: string; refresh_token: string; expires_in: number }>(
    '/miniapp/auth/refresh',
    { refresh_token: refreshToken }
  )
}
