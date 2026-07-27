import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUIStore = defineStore('ui', () => {
  const sidebarCollapsed = ref(false)
  const loading = ref(false)
  const globalLoadingText = ref('加载中...')

  function toggleSidebar() {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function setSidebarCollapsed(collapsed: boolean) {
    sidebarCollapsed.value = collapsed
  }

  function showLoading(text = '加载中...') {
    loading.value = true
    globalLoadingText.value = text
  }

  function hideLoading() {
    loading.value = false
  }

  return {
    sidebarCollapsed,
    loading,
    globalLoadingText,
    toggleSidebar,
    setSidebarCollapsed,
    showLoading,
    hideLoading
  }
})
