<template>
  <div>
    <div class="avg">
      平均评分：<RatingStars :value="avgRating" />（{{ avgRating.toFixed(1) }}）
    </div>
    <el-divider />
    <EmptyState v-if="list.length === 0" text="暂无评论" />
    <el-card v-for="c in list" :key="c.id" shadow="never" class="comment-item">
      <div class="row">
        <RatingStars :value="c.rating" />
        <span class="time">{{ formatDateTime(c.created_at) }}</span>
      </div>
      <div class="content">{{ c.content }}</div>
      <div class="actions">
        <RoleGuard :roles="['admin']">
          <el-button size="small" type="danger" text @click="remove(c.id)">删除</el-button>
        </RoleGuard>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import RatingStars from '@/components/common/RatingStars.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import RoleGuard from '@/components/common/RoleGuard.vue'
import { useCommentStore } from '@/stores/commentStore'
import { deleteComment } from '@/api/comment'
import { formatDateTime } from '@/utils/dateFormat'

const props = defineProps<{ activityId: number }>()
const store = useCommentStore()
const list = store.list
const avgRating = store.avgRating

async function remove(id: number) {
  await deleteComment(id)
  ElMessage.success('已删除')
  await store.fetchList(props.activityId)
}

onMounted(() => store.fetchList(props.activityId))
</script>

<style scoped>
.avg { display: flex; align-items: center; gap: 8px; }
.comment-item { margin-bottom: 8px; }
.row { display: flex; justify-content: space-between; }
.time { color: #909399; font-size: 12px; }
.content { margin-top: 6px; }
.actions { text-align: right; }
</style>
