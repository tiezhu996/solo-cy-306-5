import { defineStore } from 'pinia'
import { listMyNotifications, markRead, markAllRead } from '@/api/notification'
import type { NotificationItem } from '@/types'

export const useNotificationStore = defineStore('notification', {
  state: () => ({
    list: [] as NotificationItem[],
    total: 0,
  }),
  getters: {
    unreadCount: (state) => state.list.filter((n) => !n.is_read).length,
  },
  actions: {
    async fetchMine(params: { page?: number; page_size?: number } = {}) {
      const res = await listMyNotifications(params)
      this.list = res.data.list
      this.total = res.data.total
    },
    async markReadOne(id: number) {
      await markRead(id)
      const item = this.list.find((n) => n.id === id)
      if (item) item.is_read = true
    },
    async markAll() {
      await markAllRead()
      this.list.forEach((n) => (n.is_read = true))
    },
  },
})
