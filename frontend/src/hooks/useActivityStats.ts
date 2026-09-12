import { ref } from 'vue'
import { getActivityStats } from '@/api/activity'

export interface ActivityStats {
  activity_id: number
  capacity: number
  registered_count: number
  checked_in_count: number
  checkin_rate: number
}

export function useActivityStats() {
  const loading = ref(false)
  const stats = ref<ActivityStats | null>(null)

  async function load(activityId: number) {
    loading.value = true
    try {
      const res = await getActivityStats(activityId)
      stats.value = res.data
    } finally {
      loading.value = false
    }
  }

  return { loading, stats, load }
}
