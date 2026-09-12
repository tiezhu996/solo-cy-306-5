import { defineStore } from 'pinia'
import { login as apiLogin, getMe } from '@/api/user'
import type { User } from '@/types'

const TOKEN_KEY = 'gbevent_token'
const USER_KEY = 'gbevent_user'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem(TOKEN_KEY) || '',
    user: JSON.parse(localStorage.getItem(USER_KEY) || 'null') as User | null,
  }),
  getters: {
    isLoggedIn: (state) => !!state.token,
    role: (state) => state.user?.role || '',
  },
  actions: {
    async login(username: string, password: string) {
      const res = await apiLogin({ username, password })
      this.token = res.data.token
      this.user = res.data.user
      localStorage.setItem(TOKEN_KEY, this.token)
      localStorage.setItem(USER_KEY, JSON.stringify(this.user))
    },
    async fetchMe() {
      const res = await getMe()
      this.user = res.data
      localStorage.setItem(USER_KEY, JSON.stringify(this.user))
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem(USER_KEY)
    },
  },
})
