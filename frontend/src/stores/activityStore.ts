import { defineStore } from 'pinia'
import { listActivities, getActivity, getCalendar } from '@/api/activity'
import type { Activity } from '@/types'

export const useActivityStore = defineStore('activity', {
  state: () => ({
    list: [] as Activity[],
    total: 0,
    calendar: [] as Activity[],
    current: null as Activity | null,
    registeredCount: 0,
  }),
  actions: {
    async fetchList(params: { page?: number; page_size?: number; activity_type?: string; status?: string; keyword?: string } = {}) {
      const res = await listActivities(params)
      this.list = res.data.list
      this.total = res.data.total
    },
    async fetchCalendar(month?: string) {
      const res = await getCalendar(month)
      this.calendar = res.data
    },
    async fetchDetail(id: number) {
      const res = await getActivity(id)
      this.current = res.data.activity
      this.registeredCount = res.data.registered_count
      return this.current
    },
  },
})
