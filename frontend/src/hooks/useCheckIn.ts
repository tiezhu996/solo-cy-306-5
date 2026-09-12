import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { checkInByVoucher, checkInByScan } from '@/api/checkIn'

export function useCheckIn() {
  const loading = ref(false)

  async function byVoucher(activityId: number, voucher: string) {
    loading.value = true
    try {
      const res = (await checkInByVoucher(activityId, voucher)) as unknown as { message?: string; data: unknown }
      ElMessage.success(res.message || '签到成功')
      return res.data
    } finally {
      loading.value = false
    }
  }

  async function byScan(activityId: number, qrContent: string) {
    loading.value = true
    try {
      const res = (await checkInByScan(activityId, qrContent)) as unknown as { message?: string; data: unknown }
      ElMessage.success(res.message || '签到成功')
      return res.data
    } finally {
      loading.value = false
    }
  }

  return { loading, byVoucher, byScan }
}
