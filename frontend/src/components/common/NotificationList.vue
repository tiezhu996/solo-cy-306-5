<template>
  <div>
    <div class="toolbar">
      <el-button size="small" @click="store.markAll()">全部已读</el-button>
      <el-badge :value="store.unreadCount" :hidden="store.unreadCount === 0">未读</el-badge>
    </div>
    <EmptyState v-if="store.list.length === 0" text="暂无通知" />
    <el-card v-for="n in store.list" :key="n.id" shadow="never" class="item" :class="{ unread: !n.is_read }">
      <div class="row">
        <el-tag size="small">{{ n.type_text || n.notification_type }}</el-tag>
        <span class="time">{{ formatDateTime(n.created_at) }}</span>
      </div>
      <div class="title">{{ n.title }}</div>
      <div class="content">{{ n.content }}</div>
      <el-button v-if="!n.is_read" size="small" text type="primary" @click="store.markReadOne(n.id)">标记已读</el-button>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import EmptyState from '@/components/common/EmptyState.vue'
import { useNotificationStore } from '@/stores/notificationStore'
import { formatDateTime } from '@/utils/dateFormat'

const store = useNotificationStore()
onMounted(() => store.fetchMine({ page: 1, page_size: 20 }))
</script>

<style scoped>
.toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.item { margin-bottom: 8px; }
.item.unread { border-left: 3px solid #409eff; }
.row { display: flex; justify-content: space-between; }
.time { color: #909399; font-size: 12px; }
.title { font-weight: 600; margin: 6px 0; }
.content { color: #606266; }
</style>
