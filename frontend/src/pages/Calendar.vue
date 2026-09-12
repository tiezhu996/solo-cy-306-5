<template>
  <div>
    <el-radio-group v-model="view" class="mb-2">
      <el-radio-button value="calendar">日历视图</el-radio-button>
      <el-radio-button value="list">列表视图</el-radio-button>
    </el-radio-group>
    <ActivityCalendar v-if="view === 'calendar'" :activities="store.calendar" @month-change="loadMonth" />
    <div v-else>
      <ActivityCard v-for="act in store.calendar" :key="act.id" :activity="act" class="mb-2" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import ActivityCalendar from '@/components/common/ActivityCalendar.vue'
import ActivityCard from '@/components/common/ActivityCard.vue'
import { useActivityStore } from '@/stores/activityStore'

const store = useActivityStore()
const view = ref('calendar')
const month = ref('')

async function loadMonth(m: string) {
  month.value = m
  await store.fetchCalendar(m)
}

onMounted(() => {
  const now = new Date()
  loadMonth(`${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`)
})
</script>

<style scoped>
.mb-2 { margin-bottom: 8px; }
</style>
