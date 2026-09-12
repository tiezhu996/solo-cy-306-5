<template>
  <div class="activity-filter">
    <el-select v-model="model.type" placeholder="活动类型" clearable style="width: 140px" @change="emitChange">
      <el-option v-for="opt in ActivityTypeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
    </el-select>
    <el-select v-model="model.status" placeholder="活动状态" clearable style="width: 140px" @change="emitChange">
      <el-option v-for="opt in ActivityStatusOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
    </el-select>
    <el-input v-model="model.keyword" placeholder="搜索活动标题" clearable style="width: 220px" @change="emitChange" @clear="emitChange" />
    <el-button type="primary" @click="emitChange">查询</el-button>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { ActivityStatusOptions, ActivityTypeOptions } from '@/constants/activity'

const emit = defineEmits<{ (e: 'change', value: { type?: string; status?: string; keyword?: string }): void }>()
const model = reactive<{ type?: string; status?: string; keyword?: string }>({})

function emitChange() {
  emit('change', { ...model })
}

defineExpose({ model })
</script>

<style scoped>
.activity-filter { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }
</style>
