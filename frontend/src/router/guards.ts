import type { Router } from 'vue-router'
import { useAuthStore } from '@/stores/authStore'

export function setupGuards(router: Router) {
  router.beforeEach((to) => {
    const auth = useAuthStore()
    if (to.meta.requiresAuth && !auth.token) {
      return { path: '/login', query: { redirect: to.fullPath } }
    }
    const roles = (to.meta.roles as string[]) || []
    if (roles.length > 0 && (!auth.user || !roles.includes(auth.user.role))) {
      return { path: '/calendar' }
    }
    return true
  })
}
