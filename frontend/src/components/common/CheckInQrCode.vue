<template>
  <div class="qr-wrap">
    <div class="qr">{{ qrContent }}</div>
    <div class="tip">扫码时输入或扫描上方报名 ID（即二维码内容）</div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ registrationId: number }>()
const qrContent = computed(() => String(props.registrationId))
</script>

<style scoped>
.qr-wrap { text-align: center; }
.qr { width: 140px; height: 140px; margin: 0 auto; display: flex; align-items: center; justify-content: center; background: #fff; border: 1px solid #ddd; font-size: 32px; font-weight: 700; }
.tip { color: #909399; font-size: 12px; margin-top: 8px; }
</style>
