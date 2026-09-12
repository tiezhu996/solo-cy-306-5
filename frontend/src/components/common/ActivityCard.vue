<template>
  <el-card class="activity-card" shadow="hover" @click="goDetail">
    <div class="cover" v-if="activity.cover_image">
      <el-image :src="activity.cover_image" fit="cover" />
    </div>
    <div class="cover placeholder" v-else>
      <el-icon :size="32"><Calendar /></el-icon>
    </div>
    <div class="body">
      <div class="title">{{ activity.title }}</div>
      <el-tag size="small" type="primary">{{ ActivityTypeText[activity.activity_type] }}</el-tag>
      <el-tag size="small" :type="statusTagType">{{ ActivityStatusText[activity.status] }}</el-tag>
      <div class="meta">
        <el-icon><Clock /></el-icon>
        <span>{{ formatDateTime(activity.start_time) }}</span>
      </div>
      <div class="meta">
        <el-icon><Location /></el-icon>
        <span>{{ activity.location }}</span>
      </div>
      <div class="footer">
        <span class="capacity">名额 {{ activity.registered_count ?? 0 }}/{{ activity.capacity }}</span>
        <el-button type="primary" size="small">查看详情</el-button>
      </div>
    </div>
  </el-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import type { Activity } from '@/types'
import { ActivityStatus, ActivityStatusText, ActivityTypeText } from '@/constants/activity'
import { formatDateTime } from '@/utils/dateFormat'

const props = defineProps<{ activity: Activity }>()
const router = useRouter()

const statusTagType = computed(() => {
  if (props.activity.status === ActivityStatus.ENDED) return 'info'
  if (props.activity.status === ActivityStatus.DRAFT) return 'warning'
  return 'success'
})

function goDetail() {
  router.push(`/activities/${props.activity.id}`)
}
</script>

<style scoped>
.activity-card { cursor: pointer; }
.cover { height: 120px; border-radius: 6px; overflow: hidden; }
.cover.placeholder { display: flex; align-items: center; justify-content: center; background: #f0f2f5; color: #c0c4cc; }
.body { padding-top: 10px; }
.title { font-size: 16px; font-weight: 600; margin-bottom: 6px; }
.meta { display: flex; align-items: center; gap: 4px; color: #909399; font-size: 13px; margin-top: 4px; }
.footer { display: flex; justify-content: space-between; align-items: center; margin-top: 10px; }
.capacity { color: #606266; font-size: 13px; }
</style>
