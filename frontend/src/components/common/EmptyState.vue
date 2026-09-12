<template>
  <el-empty :description="text">
    <slot />
  </el-empty>
</template>

<script setup lang="ts">
defineProps<{ text?: string }>()
</script>
