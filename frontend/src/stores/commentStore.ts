import { defineStore } from 'pinia'
import { listComments, listMyComments } from '@/api/comment'
import type { CommentItem } from '@/types'

export const useCommentStore = defineStore('comment', {
  state: () => ({
    list: [] as CommentItem[],
    mine: [] as CommentItem[],
    avgRating: 0,
  }),
  actions: {
    async fetchList(activityId: number) {
      const res = await listComments(activityId)
      this.list = res.data.list
      this.avgRating = res.data.avg_rating
    },
    async fetchMine() {
      const res = await listMyComments()
      this.mine = res.data
    },
  },
})
