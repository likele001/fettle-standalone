const TOKEN_KEY = 'platform_access_token'
const REFRESH_KEY = 'platform_refresh_token'
const EXPIRE_KEY = 'platform_token_expires_at'

export function getAccessToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_KEY)
}

export function setTokens(accessToken: string, refreshToken: string, expiresIn: number) {
  localStorage.setItem(TOKEN_KEY, accessToken)
  localStorage.setItem(REFRESH_KEY, refreshToken)
  const expiresAt = Date.now() + expiresIn * 1000
  localStorage.setItem(EXPIRE_KEY, expiresAt.toString())
}

export function removeTokens() {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(REFRESH_KEY)
  localStorage.removeItem(EXPIRE_KEY)
}

export function isTokenExpired(): boolean {
  const expiresAt = localStorage.getItem(EXPIRE_KEY)
  if (!expiresAt) return true
  return Date.now() > parseInt(expiresAt) - 5 * 60 * 1000
}
