import { computed } from 'vue'
import { useAuthStore } from '@/stores/authStore'
import { UserRole } from '@/constants/notification'

export function useAuth() {
  const auth = useAuthStore()
  const isLoggedIn = computed(() => auth.isLoggedIn)
  const role = computed(() => auth.role)
  const isOrganizer = computed(() => role.value === UserRole.ORGANIZER || role.value === UserRole.ADMIN)
  const isAdmin = computed(() => role.value === UserRole.ADMIN)
  const hasRole = (...roles: string[]) => roles.includes(role.value)
  return { auth, isLoggedIn, role, isOrganizer, isAdmin, hasRole }
}
