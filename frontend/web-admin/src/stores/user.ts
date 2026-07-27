import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { UserInfo, TokenResponse } from '@/api/auth'
import { isTokenExpired, setToken, setRefreshToken, clearTokens } from '@/utils/token'

export const useUserStore = defineStore('user', () => {
  const accessToken = ref<string>(localStorage.getItem('access_token') || '')
  const refreshToken = ref<string>(localStorage.getItem('refresh_token') || '')
  const userInfo = ref<UserInfo | null>(
    localStorage.getItem('user_info') ? JSON.parse(localStorage.getItem('user_info')!) : null
  )

  const isLoggedIn = computed(() => !!accessToken.value)
  const isTokenExpiredComputed = computed(() => isTokenExpired())

  function setTokens(tokens: TokenResponse) {
    accessToken.value = tokens.access_token
    refreshToken.value = tokens.refresh_token
    setToken(tokens.access_token, tokens.expires_in || 900)
    setRefreshToken(tokens.refresh_token)
  }

  function setUser(info: UserInfo) {
    userInfo.value = info
    localStorage.setItem('user_info', JSON.stringify(info))
  }

  function clearUser() {
    accessToken.value = ''
    refreshToken.value = ''
    userInfo.value = null
    clearTokens()
    localStorage.removeItem('user_info')
  }

  return {
    accessToken,
    refreshToken,
    userInfo,
    isLoggedIn,
    isTokenExpired: isTokenExpiredComputed,
    setTokens,
    setUser,
    clearUser
  }
})
