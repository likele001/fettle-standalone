import type { Router } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { getToken } from '@/utils/token'
import { ElMessage } from 'element-plus'

export function setupRouterGuards(router: Router) {
  router.beforeEach((to, from, next) => {
    const userStore = useUserStore()
    
    if (to.meta.public) {
      next()
      return
    }

    if (to.meta.requiresAuth) {
      const hasToken = !!getToken() || !!userStore.accessToken
      if (!hasToken) {
        ElMessage.warning('请先登录')
        next({ path: '/login', query: { redirect: to.fullPath } })
        return
      }
    }

    next()
  })

  router.afterEach((to) => {
    if (to.meta.title) {
      document.title = to.meta.title + ' - AI智能体平台'
    }
  })
}
