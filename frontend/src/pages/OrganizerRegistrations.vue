<template>
  <div>
    <el-form inline class="toolbar">
      <el-form-item label="活动">
        <el-select v-model="activityId" placeholder="选择活动" clearable style="width: 220px" @change="load">
          <el-option v-for="a in activities" :key="a.id" :label="a.title" :value="a.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="statusFilter" placeholder="状态" clearable style="width: 130px" @change="load">
          <el-option v-for="(text, val) in RegistrationStatusText" :key="val" :label="text" :value="val" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button v-if="activityId" type="success" @click="exportCsv">导出名单</el-button>
        <el-button type="primary" @click="offlineVisible = true">线下补录</el-button>
      </el-form-item>
    </el-form>

    <el-dialog v-model="offlineVisible" title="线下补录" width="460px">
      <el-form :model="offlineForm" label-width="80px">
        <el-form-item label="姓名"><el-input v-model="offlineForm.name" /></el-form-item>
        <el-form-item label="手机号"><el-input v-model="offlineForm.phone" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="offlineForm.remark" /></el-form-item>
        <el-button type="primary" @click="submitOffline">提交</el-button>
      </el-form>
    </el-dialog>

    <el-tabs>
      <el-tab-pane label="报名名单">
        <RegistrationTable
          :list="store.list"
          :total="store.total"
          :page="page"
          :page-size="pageSize"
          :loading="loading"
          @reviewed="load"
          @checkin="openCheckin"
          @page-change="onPage"
        />
      </el-tab-pane>
      <el-tab-pane label="签到面板" v-if="activityId">
        <CheckInPanel :activity-id="activityId" />
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="checkinVisible" title="凭证签到" width="420px">
      <CheckInQrCode :registration-id="checkinRow?.id || 0" />
      <el-divider />
      <el-input v-model="voucherInput" placeholder="输入凭证号" />
      <el-button type="primary" class="mt-2" @click="doCheckin">确认签到</el-button>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import RegistrationTable from '@/components/common/RegistrationTable.vue'
import CheckInPanel from '@/components/common/CheckInPanel.vue'
import CheckInQrCode from '@/components/common/CheckInQrCode.vue'
import { useRegistrationStore } from '@/stores/registrationStore'
import { offlineSignup, exportRegistrations } from '@/api/registration'
import { useCheckIn } from '@/hooks/useCheckIn'
import { RegistrationStatusText } from '@/constants/registration'
import type { Activity, Registration } from '@/types'

const store = useRegistrationStore()
const { byVoucher } = useCheckIn()
const loading = ref(false)
const page = ref(1)
const pageSize = 10
const activityId = ref<number>()
const statusFilter = ref('')
const activities = ref<Activity[]>([])
const offlineVisible = ref(false)
const offlineForm = ref({ name: '', phone: '', remark: '' })
const checkinVisible = ref(false)
const checkinRow = ref<Registration | null>(null)
const voucherInput = ref('')

async function loadActivities() {
  const res = await fetch('/api/v1/activities/mine?page=1&page_size=100', {
    headers: { Authorization: `Bearer ${localStorage.getItem('gbevent_token')}` },
  })
  const body = await res.json()
  activities.value = body.data.list
  if (!activityId.value && activities.value.length > 0) {
    activityId.value = activities.value[0].id
  }
}

async function load() {
  loading.value = true
  try {
    await store.fetchList({ page: page.value, page_size: pageSize, activity_id: activityId.value, status: statusFilter.value || undefined })
  } finally {
    loading.value = false
  }
}

function onPage(p: number) {
  page.value = p
  load()
}

function openCheckin(row: Registration) {
  checkinRow.value = row
  voucherInput.value = row.voucher_no
  checkinVisible.value = true
}

async function doCheckin() {
  if (!activityId.value) return
  await byVoucher(activityId.value, voucherInput.value)
  checkinVisible.value = false
  load()
}

async function submitOffline() {
  if (!activityId.value) return
  await offlineSignup({ activity_id: activityId.value, ...offlineForm.value })
  ElMessage.success('补录成功')
  offlineVisible.value = false
  offlineForm.value = { name: '', phone: '', remark: '' }
  load()
}

async function exportCsv() {
  if (!activityId.value) return
  const res = await exportRegistrations(activityId.value)
  const blob = new Blob([res.data])
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `registrations_${activityId.value}.csv`
  a.click()
  URL.revokeObjectURL(url)
}

onMounted(async () => {
  await loadActivities()
  await load()
})
</script>

<style scoped>
.toolbar { margin-bottom: 12px; }
.mt-2 { margin-top: 8px; }
</style>
