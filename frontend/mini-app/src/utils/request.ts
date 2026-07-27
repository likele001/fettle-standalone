import { getToken } from './token'

const BASE_URL = import.meta.env.VITE_API_BASE_URL || 'https://api.fettle.cenkor.cn/v1'

interface RequestOptions {
  url: string
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  data?: any
  header?: Record<string, string>
  showLoading?: boolean
  showError?: boolean
}

interface Response<T = any> {
  code: number
  message: string
  data: T
}

export function request<T = any>(options: RequestOptions): Promise<T> {
  return new Promise((resolve, reject) => {
    const token = getToken()
    const header: Record<string, string> = {
      'Content-Type': 'application/json',
      ...options.header
    }

    if (token) {
      header['Authorization'] = `Bearer ${token}`
    }

    if (options.showLoading !== false) {
      uni.showLoading({
        title: '加载中...',
        mask: true
      })
    }

    uni.request({
      url: BASE_URL + options.url,
      method: options.method || 'GET',
      data: options.data,
      header,
      success: (res) => {
        if (options.showLoading !== false) {
          uni.hideLoading()
        }

        const response = res.data as Response<T>
        
        if (response.code === 0) {
          resolve(response.data)
        } else {
          if (options.showError !== false) {
            uni.showToast({
              title: response.message || '请求失败',
              icon: 'none',
              duration: 2000
            })
          }
          reject(new Error(response.message))
        }
      },
      fail: (err) => {
        if (options.showLoading !== false) {
          uni.hideLoading()
        }

        if (options.showError !== false) {
          uni.showToast({
            title: '网络错误，请稍后重试',
            icon: 'none',
            duration: 2000
          })
        }
        
        reject(err)
      }
    })
  })
}

// 便捷方法
export function get<T = any>(url: string, data?: any, options?: Partial<RequestOptions>): Promise<T> {
  return request<T>({ url, method: 'GET', data, ...options })
}

export function post<T = any>(url: string, data?: any, options?: Partial<RequestOptions>): Promise<T> {
  return request<T>({ url, method: 'POST', data, ...options })
}

export function put<T = any>(url: string, data?: any, options?: Partial<RequestOptions>): Promise<T> {
  return request<T>({ url, method: 'PUT', data, ...options })
}

export function del<T = any>(url: string, data?: any, options?: Partial<RequestOptions>): Promise<T> {
  return request<T>({ url, method: 'DELETE', data, ...options })
}
