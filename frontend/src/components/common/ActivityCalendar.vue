<template>
  <div>
    <el-calendar v-model="currentDate">
      <template #date-cell="{ data }">
        <div class="calendar-cell" @click="selectDate(data.day)">
          <span>{{ data.day.split('-')[2] }}</span>
          <div v-if="countByDate(data.day) > 0" class="count">{{ countByDate(data.day) }} 场</div>
        </div>
      </template>
    </el-calendar>
    <el-divider>当天活动</el-divider>
    <div v-if="dayActivities.length === 0" class="empty">当日暂无活动</div>
    <ActivityCard v-for="act in dayActivities" :key="act.id" :activity="act" class="mb-2" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import ActivityCard from '@/components/common/ActivityCard.vue'
import type { Activity } from '@/types'
import { formatDate } from '@/utils/dateFormat'

const props = defineProps<{ activities: Activity[]; month?: string }>()
const emit = defineEmits<{ (e: 'month-change', month: string): void }>()

const currentDate = ref(new Date())

watch(currentDate, (d) => {
  emit('month-change', `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`)
}, { immediate: true })

function countByDate(day: string): number {
  return props.activities.filter((a) => formatDate(a.start_time) === day).length
}

const selectedDate = ref(formatDate(new Date()))
function selectDate(day: string) {
  selectedDate.value = day
}
const dayActivities = computed(() => props.activities.filter((a) => formatDate(a.start_time) === selectedDate.value))
</script>

<style scoped>
.calendar-cell { cursor: pointer; min-height: 60px; }
.count { color: #409eff; font-size: 12px; margin-top: 4px; }
.empty { color: #909399; text-align: center; padding: 16px 0; }
.mb-2 { margin-bottom: 8px; }
</style>
