const TOKEN_KEY = 'access_token'
const REFRESH_TOKEN_KEY = 'refresh_token'
const TOKEN_EXPIRES_KEY = 'token_expires_at'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string, expiresIn: number) {
  localStorage.setItem(TOKEN_KEY, token)
  // 防御性检查：确保 expiresIn 有效，默认 900 秒（15分钟）
  const validExpiresIn = (expiresIn && expiresIn > 0) ? expiresIn : 900
  const expiresAt = Date.now() + validExpiresIn * 1000
  localStorage.setItem(TOKEN_EXPIRES_KEY, expiresAt.toString())
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

export function setRefreshToken(token: string) {
  localStorage.setItem(REFRESH_TOKEN_KEY, token)
}

export function clearTokens() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
  localStorage.removeItem(TOKEN_EXPIRES_KEY)
}

export function isTokenExpired(): boolean {
  const expiresAt = localStorage.getItem(TOKEN_EXPIRES_KEY)
  if (!expiresAt) return true

  const now = Date.now()
  const expires = parseInt(expiresAt, 10)

  // 提前5分钟认为过期，留出刷新时间
  return now >= expires - 5 * 60 * 1000
}

export function getTokenRemainingTime(): number {
  const expiresAt = localStorage.getItem(TOKEN_EXPIRES_KEY)
  if (!expiresAt) return 0

  const now = Date.now()
  const expires = parseInt(expiresAt, 10)

  return Math.max(0, expires - now)
}
