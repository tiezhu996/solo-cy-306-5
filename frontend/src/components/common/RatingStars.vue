<template>
  <el-rate :model-value="value" disabled />
</template>

<script setup lang="ts">
defineProps<{ value: number }>()
</script>
