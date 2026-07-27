import { getToken, isTokenExpired, clearTokens } from './token'

export function isAuthenticated(): boolean {
  const token = getToken()
  if (!token) return false
  if (isTokenExpired()) {
    clearTokens()
    return false
  }
  return true
}

export function getAuthHeader(): Record<string, string> {
  const token = getToken()
  if (!token) return {}
  return {
    Authorization: `Bearer ${token}`
  }
}

export function logout() {
  clearTokens()
  localStorage.removeItem('user_info')
}

export function getUserInfo(): any {
  const userInfo = localStorage.getItem('user_info')
  if (!userInfo) return null
  try {
    return JSON.parse(userInfo)
  } catch {
    return null
  }
}

export function setUserInfo(userInfo: any) {
  localStorage.setItem('user_info', JSON.stringify(userInfo))
}

export function getUserRole(): string {
  const userInfo = getUserInfo()
  return userInfo?.role || ''
}

export function hasRole(requiredRole: string): boolean {
  const roleHierarchy: Record<string, number> = {
    super_admin: 5,
    admin: 4,
    operator: 3,
    member: 2,
    viewer: 1
  }

  const userRole = getUserRole()
  const userLevel = roleHierarchy[userRole] || 0
  const requiredLevel = roleHierarchy[requiredRole] || 0

  return userLevel >= requiredLevel
}
