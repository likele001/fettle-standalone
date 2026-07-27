import type { Router, RouteLocationNormalized } from 'vue-router'
import { useUserStore } from '@/stores/user'

export function setupGuards(router: Router) {
  router.beforeEach((to: RouteLocationNormalized) => {
    const userStore = useUserStore()

    if (to.meta.public) {
      if (userStore.isLoggedIn && !userStore.isTokenExpired()) {
        return '/'
      }
      return true
    }

    if (!userStore.isLoggedIn || userStore.isTokenExpired()) {
      userStore.clearUser()
      return '/login'
    }

    if (!userStore.isSuperAdmin) {
      userStore.clearUser()
      return '/login'
    }

    return true
  })
}
