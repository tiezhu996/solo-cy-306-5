import { defineStore } from 'pinia'
import { listRegistrations, listMyRegistrations } from '@/api/registration'
import type { PageResult, Registration } from '@/types'

export const useRegistrationStore = defineStore('registration', {
  state: () => ({
    list: [] as Registration[],
    mine: [] as Registration[],
    total: 0,
    mineTotal: 0,
  }),
  actions: {
    async fetchList(params: { page?: number; page_size?: number; activity_id?: number; status?: string; review_status?: string } = {}) {
      const res = await listRegistrations(params)
      const data = res.data as PageResult<Registration>
      this.list = data.list
      this.total = data.total
    },
    async fetchMine(params: { page?: number; page_size?: number } = {}) {
      const res = await listMyRegistrations(params)
      const data = res.data as PageResult<Registration>
      this.mine = data.list
      this.mineTotal = data.total
    },
  },
})
