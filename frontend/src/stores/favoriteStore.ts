import { defineStore } from 'pinia'
import { listMyFavorites } from '@/api/favorite'
import type { Favorite } from '@/types'

export const useFavoriteStore = defineStore('favorite', {
  state: () => ({
    list: [] as Favorite[],
    total: 0,
  }),
  actions: {
    async fetchMine(params: { page?: number; page_size?: number } = {}) {
      const res = await listMyFavorites(params)
      this.list = res.data.list
      this.total = res.data.total
    },
  },
})
