<template>
  <slot v-if="allowed" />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuth } from '@/hooks/useAuth'

const props = defineProps<{ roles: string[] }>()
const { hasRole } = useAuth()
const allowed = computed(() => hasRole(...props.roles))
</script>
