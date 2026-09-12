<template>
  <div>
    <div class="toolbar">
      <el-button type="primary" @click="openCreate">创建活动</el-button>
      <el-select v-model="statusFilter" placeholder="状态" clearable style="width: 140px" @change="load">
        <el-option v-for="opt in ActivityStatusOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
      </el-select>
    </div>

    <el-dialog v-model="dialogVisible" :title="editing ? '编辑活动' : '创建活动'" width="640px">
      <ActivityForm :activity="editing" @success="onSaved" />
    </el-dialog>

    <el-table :data="store.list" v-loading="loading" border>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="title" label="标题" min-width="200" />
      <el-table-column label="类型" width="100">
        <template #default="{ row }">{{ ActivityTypeText[row.activity_type] }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'published' ? 'success' : row.status === 'draft' ? 'warning' : 'info'">
            {{ ActivityStatusText[row.status] }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="location" label="地点" min-width="140" />
      <el-table-column label="操作" min-width="260">
        <template #default="{ row }">
          <el-button size="small" @click="edit(row)">编辑</el-button>
          <el-button v-if="row.status === 'draft'" size="small" type="success" @click="publish(row)">发布</el-button>
          <el-button v-if="row.status === 'published'" size="small" type="warning" @click="end(row)">结束</el-button>
          <el-button size="small" type="danger" @click="remove(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="mt-2" layout="total, prev, pager, next" :total="store.total" :page-size="pageSize" :current-page="page" @current-change="onPage" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import ActivityForm from '@/components/common/ActivityForm.vue'
import { useActivityStore } from '@/stores/activityStore'
import { publishActivity, endActivity, deleteActivity } from '@/api/activity'
import { ActivityStatusOptions, ActivityStatusText, ActivityTypeText } from '@/constants/activity'
import type { Activity } from '@/types'

const store = useActivityStore()
const loading = ref(false)
const page = ref(1)
const pageSize = 10
const statusFilter = ref('')
const dialogVisible = ref(false)
const editing = ref<Activity | null>(null)

async function load() {
  loading.value = true
  try {
    await store.fetchList({ page: page.value, page_size: pageSize })
    const res = await fetch(`/api/v1/activities/mine?page=${page.value}&page_size=${pageSize}&status=${statusFilter.value}`, {
      headers: { Authorization: `Bearer ${localStorage.getItem('gbevent_token')}` },
    })
    const body = await res.json()
    store.list = body.data.list
    store.total = body.data.total
  } finally {
    loading.value = false
  }
}
function openCreate() {
  editing.value = null
  dialogVisible.value = true
}
function edit(row: Activity) {
  editing.value = row
  dialogVisible.value = true
}
function onSaved() {
  dialogVisible.value = false
  load()
}
async function publish(row: Activity) {
  await publishActivity(row.id)
  ElMessage.success('已发布')
  load()
}
async function end(row: Activity) {
  await endActivity(row.id)
  ElMessage.success('已结束')
  load()
}
async function remove(row: Activity) {
  await ElMessageBox.confirm('确认删除该活动？', '提示')
  await deleteActivity(row.id)
  ElMessage.success('已删除')
  load()
}
function onPage(p: number) {
  page.value = p
  load()
}
onMounted(load)
</script>

<style scoped>
.toolbar { display: flex; gap: 12px; margin-bottom: 12px; }
.mt-2 { margin-top: 12px; }
</style>
