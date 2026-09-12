<template>
  <div>
    <ActivityFilter @change="applyFilter" />
    <div v-loading="loading" class="grid">
      <ActivityCard v-for="act in store.list" :key="act.id" :activity="act" />
      <EmptyState v-if="!loading && store.list.length === 0" text="暂无活动" />
    </div>
    <el-pagination
      class="mt-2"
      layout="total, prev, pager, next"
      :total="store.total"
      :page-size="pageSize"
      :current-page="page"
      @current-change="onPage"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import ActivityFilter from '@/components/common/ActivityFilter.vue'
import ActivityCard from '@/components/common/ActivityCard.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { useActivityStore } from '@/stores/activityStore'

const store = useActivityStore()
const page = ref(1)
const pageSize = 12
const loading = ref(false)
const filter = ref<{ type?: string; status?: string; keyword?: string }>({})

async function load() {
  loading.value = true
  try {
    await store.fetchList({ page: page.value, page_size: pageSize, ...filter.value })
  } finally {
    loading.value = false
  }
}
function applyFilter(f: { type?: string; status?: string; keyword?: string }) {
  filter.value = f
  page.value = 1
  load()
}
function onPage(p: number) {
  page.value = p
  load()
}
onMounted(load)
</script>

<style scoped>
.grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 16px; margin-top: 16px; }
.mt-2 { margin-top: 16px; }
</style>
