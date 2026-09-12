<template>
  <el-table :data="list" v-loading="loading" border>
    <el-table-column prop="id" label="ID" width="70" />
    <el-table-column prop="activity_id" label="活动ID" width="90" />
    <el-table-column prop="voucher_no" label="凭证号" width="180" />
    <el-table-column label="状态" width="100">
      <template #default="{ row }">
        <el-tag :type="tag(row.status)">{{ RegistrationStatusText[row.status] }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="审核" width="100">
      <template #default="{ row }">
        <el-tag :type="reviewTag(row.review_status)">{{ ReviewStatusText[row.review_status] }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="created_at" label="报名时间" />
    <el-table-column label="操作" width="100">
      <template #default="{ row }">
        <el-button v-if="row.status === 'registered'" size="small" type="danger" @click="cancel(row)">取消</el-button>
      </template>
    </el-table-column>
  </el-table>
  <el-pagination
    class="mt-2"
    layout="total, prev, pager, next"
    :total="total"
    :page-size="pageSize"
    @current-change="onPage"
  />
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { cancelRegistration } from '@/api/registration'
import { RegistrationStatusText, ReviewStatusText } from '@/constants/registration'
import type { Registration } from '@/types'

const list = ref<Registration[]>([])
const total = ref(0)
const loading = ref(false)
const page = ref(1)
const pageSize = 10

async function load() {
  loading.value = true
  try {
    const res = await fetch(`/api/v1/registrations/mine?page=${page.value}&page_size=${pageSize}`, {
      headers: { Authorization: `Bearer ${localStorage.getItem('gbevent_token')}` },
    })
    const body = await res.json()
    list.value = body.data.list
    total.value = body.data.total
  } finally {
    loading.value = false
  }
}

async function cancel(row: Registration) {
  await ElMessageBox.confirm('确认取消该报名？', '提示')
  await cancelRegistration(row.id)
  ElMessage.success('已取消')
  await load()
}

function onPage(p: number) {
  page.value = p
  load()
}
function tag(status: string): string {
  if (status === 'checked_in') return 'success'
  if (status === 'cancelled') return 'info'
  return 'primary'
}
function reviewTag(status: string): string {
  if (status === 'approved') return 'success'
  if (status === 'rejected') return 'danger'
  return 'warning'
}

onMounted(load)
</script>

<style scoped>
.mt-2 { margin-top: 12px; }
</style>
