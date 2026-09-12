<template>
  <div v-loading="loading">
    <EmptyState v-if="items.length === 0" text="暂无收藏" />
    <el-card v-for="item in items" :key="item.id" shadow="never" class="item">
      <div class="row">
        <span>收藏了活动 #{{ item.activity_id }}</span>
        <span class="time">{{ formatDateTime(item.created_at) }}</span>
      </div>
      <el-button size="small" @click="goDetail(item.activity_id)">查看活动</el-button>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import EmptyState from '@/components/common/EmptyState.vue'
import { listMyFavorites } from '@/api/favorite'
import type { Favorite } from '@/types'
import { formatDateTime } from '@/utils/dateFormat'

const items = ref<Favorite[]>([])
const loading = ref(false)
const router = useRouter()

async function load() {
  loading.value = true
  try {
    const res = await listMyFavorites({ page: 1, page_size: 20 })
    items.value = res.data.list
  } finally {
    loading.value = false
  }
}
function goDetail(id: number) {
  router.push(`/activities/${id}`)
}
onMounted(load)
</script>

<style scoped>
.item { margin-bottom: 8px; }
.row { display: flex; justify-content: space-between; }
.time { color: #909399; font-size: 12px; }
</style>
