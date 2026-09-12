<template>
  <div>
    <el-tabs>
      <el-tab-pane label="凭证签到">
        <el-input v-model="voucher" placeholder="输入报名凭证号" style="width: 280px" />
        <el-button type="primary" :loading="loading" @click="doVoucher">签到</el-button>
      </el-tab-pane>
      <el-tab-pane label="扫码签到">
        <el-input v-model="qrContent" placeholder="输入扫码内容（报名 ID）" style="width: 280px" />
        <el-button type="primary" :loading="loading" @click="doScan">签到</el-button>
      </el-tab-pane>
    </el-tabs>
    <el-divider>签到记录</el-divider>
    <el-table :data="records" border size="small">
      <el-table-column prop="registration_id" label="报名ID" width="90" />
      <el-table-column label="方式" width="100">
        <template #default="{ row }">{{ CheckInMethodText[row.check_in_method] }}</template>
      </el-table-column>
      <el-table-column prop="check_in_time" label="签到时间" />
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useCheckIn } from '@/hooks/useCheckIn'
import { listCheckInRecords } from '@/api/checkIn'
import { CheckInMethodText } from '@/constants/registration'
import type { CheckInRecord } from '@/types'

const props = defineProps<{ activityId: number }>()
const { loading, byVoucher, byScan } = useCheckIn()
const voucher = ref('')
const qrContent = ref('')
const records = ref<CheckInRecord[]>([])

async function loadRecords() {
  const res = await listCheckInRecords(props.activityId)
  records.value = res.data
}

async function doVoucher() {
  await byVoucher(props.activityId, voucher.value)
  voucher.value = ''
  await loadRecords()
}

async function doScan() {
  await byScan(props.activityId, qrContent.value)
  qrContent.value = ''
  await loadRecords()
}

onMounted(loadRecords)
</script>
