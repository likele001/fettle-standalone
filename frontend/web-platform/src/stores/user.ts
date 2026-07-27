import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import { getAccessToken, getRefreshToken, setTokens as saveTokens, removeTokens, isTokenExpired as checkTokenExpired } from '@/utils/token'

interface UserInfo {
  id: string
  phone: string
  name: string
  role: string
  email?: string
  avatar_url?: string
}

export const useUserStore = defineStore('user', () => {
  const accessToken = ref<string | null>(getAccessToken())
  const refreshToken = ref<string | null>(getRefreshToken())
  const userInfo = ref<UserInfo | null>(null)

  const stored = localStorage.getItem('platform_user')
  if (stored) {
    try { userInfo.value = JSON.parse(stored) } catch { /* ignore */ }
  }

  const isLoggedIn = computed(() => !!accessToken.value)
  const isSuperAdmin = computed(() => userInfo.value?.role === 'super_admin')

  function setTokens(tokens: { access_token: string; refresh_token: string; expires_in: number }) {
    accessToken.value = tokens.access_token
    refreshToken.value = tokens.refresh_token
    saveTokens(tokens.access_token, tokens.refresh_token, tokens.expires_in)
  }

  function setUser(info: UserInfo) {
    userInfo.value = info
    localStorage.setItem('platform_user', JSON.stringify(info))
  }

  function clearUser() {
    accessToken.value = null
    refreshToken.value = null
    userInfo.value = null
    removeTokens()
    localStorage.removeItem('platform_user')
  }

  function isTokenExpired() {
    return checkTokenExpired()
  }

  return { accessToken, refreshToken, userInfo, isLoggedIn, isSuperAdmin, isTokenExpired, setTokens, setUser, clearUser }
})
