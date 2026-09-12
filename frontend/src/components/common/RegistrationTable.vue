<template>
  <div>
    <el-table :data="list" v-loading="loading" border>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="姓名" width="110" />
      <el-table-column prop="phone" label="手机号" width="130" />
      <el-table-column prop="voucher_no" label="凭证号" width="170" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTag(row.status)">{{ RegistrationStatusText[row.status] }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="审核" width="100">
        <template #default="{ row }">
          <el-tag :type="reviewTag(row.review_status)">{{ ReviewStatusText[row.review_status] }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="报名时间" width="170" />
      <el-table-column label="操作" min-width="220">
        <template #default="{ row }">
          <el-button v-if="row.review_status === 'pending'" size="small" type="success" @click="review(row, 'approved')">通过</el-button>
          <el-button v-if="row.review_status === 'pending'" size="small" type="danger" @click="review(row, 'rejected')">拒绝</el-button>
          <el-button size="small" type="primary" @click="emit('checkin', row)">签到</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      class="mt-2"
      layout="total, prev, pager, next"
      :total="total"
      :page-size="pageSize"
      :current-page="page"
      @current-change="(p: number) => emit('page-change', p)"
    />
  </div>
</template>

<script setup lang="ts">
import { reviewRegistration } from '@/api/registration'
import { RegistrationStatusText, ReviewStatusText } from '@/constants/registration'
import { ElMessage } from 'element-plus'
import type { Registration } from '@/types'

const props = defineProps<{
  list: Registration[]
  total: number
  page: number
  pageSize: number
  loading?: boolean
}>()

const emit = defineEmits<{
  (e: 'reviewed'): void
  (e: 'checkin', row: Registration): void
  (e: 'page-change', page: number): void
}>()

function statusTag(status: string): string {
  if (status === 'checked_in') return 'success'
  if (status === 'cancelled') return 'info'
  return 'primary'
}
function reviewTag(status: string): string {
  if (status === 'approved') return 'success'
  if (status === 'rejected') return 'danger'
  return 'warning'
}
async function review(row: Registration, status: string) {
  await reviewRegistration(row.id, status)
  ElMessage.success('审核完成')
  emit('reviewed')
}
</script>

<style scoped>
.mt-2 { margin-top: 12px; }
</style>
