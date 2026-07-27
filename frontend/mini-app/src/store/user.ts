import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getMiniappUserInfo } from '@/api/auth'
import { getToken, setToken, removeToken } from '@/utils/token'

export interface UserInfo {
  id: string
  name: string
  phone: string
  email: string
  avatar_url: string
  role: string
  tenant_id: string
}

export const useUserStore = defineStore('user', () => {
  const token = ref<string>(getToken() || '')
  const userInfo = ref<UserInfo | null>(null)

  const isLoggedIn = computed(() => !!token.value)
  const userName = computed(() => userInfo.value?.name || '')
  const userRole = computed(() => userInfo.value?.role || '')

  async function saveLogin(t: string) {
    token.value = t
    setToken(t)
  }

  async function fetchUser() {
    if (!token.value) return
    try {
      const res = await getMiniappUserInfo()
      userInfo.value = res
      return res
    } catch {
      // Token 失效
      logout()
      throw new Error('登录已过期，请重新登录')
    }
  }

  function logout() {
    token.value = ''
    userInfo.value = null
    removeToken()
    uni.reLaunch({ url: '/pages/login/index' })
  }

  function checkLoginStatus() {
    const savedToken = getToken()
    if (savedToken) {
      token.value = savedToken
      fetchUser().catch(() => {
        // Token 失效，清除登录状态
        logout()
      })
    }
  }

  return {
    token,
    userInfo,
    isLoggedIn,
    userName,
    userRole,
    saveLogin,
    fetchUser,
    logout,
    checkLoginStatus
  }
})
