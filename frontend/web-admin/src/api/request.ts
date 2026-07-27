import axios, { type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { refreshToken as refreshTokenApi } from '@/api/auth'
import { getToken, getRefreshToken } from '@/utils/token'

const baseURL = import.meta.env.VITE_API_BASE_URL || '/api'

const request = axios.create({
  baseURL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

let isRefreshing = false
let pendingRequests: Array<(token: string) => void> = []

function onRefreshed(token: string) {
  pendingRequests.forEach(callback => callback(token))
  pendingRequests = []
}

function addPendingRequest(callback: (token: string) => void) {
  pendingRequests.push(callback)
}

request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = getToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  (response: AxiosResponse) => {
    const data = response.data
    if (data.code !== undefined && data.code !== 0) {
      ElMessage.error(data.message || '请求失败')
      return Promise.reject(new Error(data.message || '请求失败'))
    }
    return data.data || data
  },
  async (error) => {
    const originalRequest = error.config
    const userStore = useUserStore()
    const status = error.response?.status

    if (status === 401 && !originalRequest._retry) {
      const storedRefreshToken = getRefreshToken()

      if (storedRefreshToken && !isRefreshing) {
        originalRequest._retry = true
        isRefreshing = true
        try {
          const authResp = await refreshTokenApi(storedRefreshToken)
          const tokens = authResp.tokens
          userStore.setTokens(tokens)
          isRefreshing = false
          onRefreshed(tokens.access_token)
          originalRequest.headers.Authorization = `Bearer ${tokens.access_token}`
          return request(originalRequest)
        } catch {
          isRefreshing = false
          pendingRequests = []
          userStore.clearUser()
          window.location.href = '/login'
          return Promise.reject(error)
        }
      }

      if (isRefreshing) {
        return new Promise((resolve) => {
          addPendingRequest((token: string) => {
            originalRequest.headers.Authorization = `Bearer ${token}`
            resolve(request(originalRequest))
          })
        })
      }

      userStore.clearUser()
      window.location.href = '/login'
    }

    const message = error.response?.data?.message || error.message || '网络错误'
    if (status !== 404) {
      ElMessage.error(message)
    }
    return Promise.reject(new Error(message))
  }
)

export default request