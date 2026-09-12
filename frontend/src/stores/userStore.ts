import { defineStore } from 'pinia'
import { getMe, updateProfile, listUsers } from '@/api/user'
import type { PageResult, User } from '@/types'

export const useUserStore = defineStore('user', {
  state: () => ({
    users: [] as User[],
    total: 0,
  }),
  actions: {
    async fetchMe() {
      const res = await getMe()
      return res.data as User
    },
    async updateProfile(data: Partial<User>) {
      const res = await updateProfile(data)
      return res.data as User
    },
    async fetchUsers(params: { page?: number; page_size?: number } = {}) {
      const res = await listUsers(params)
      const data = res.data as PageResult<User>
      this.users = data.list
      this.total = data.total
    },
  },
})
