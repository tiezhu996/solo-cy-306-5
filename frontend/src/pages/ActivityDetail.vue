<template>
  <div v-loading="loading">
    <template v-if="activity">
      <h2>{{ activity.title }}</h2>
      <div class="meta">
        <el-tag>{{ ActivityTypeText[activity.activity_type] }}</el-tag>
        <el-tag :type="activity.status === 'published' ? 'success' : 'info'">{{ ActivityStatusText[activity.status] }}</el-tag>
        <span>时间：{{ formatDateTime(activity.start_time) }} ~ {{ formatDateTime(activity.end_time) }}</span>
        <span>地点：{{ activity.location }}</span>
        <span>名额：{{ registeredCount }}/{{ activity.capacity }}</span>
      </div>
      <el-divider />
      <p class="desc">{{ activity.description }}</p>

      <el-tabs>
        <el-tab-pane label="在线报名">
          <template v-if="auth.isLoggedIn">
            <SignupForm v-if="canSignup" :activity-id="activity.id" @success="onSignup" />
            <el-alert v-else title="该活动当前不可报名" type="info" :closable="false" />
          </template>
          <el-alert v-else title="请先登录后报名" type="warning" :closable="false" />
        </el-tab-pane>
        <el-tab-pane label="评论">
          <el-form v-if="auth.isLoggedIn" :model="commentForm" class="mb-2">
            <el-form-item label="评分">
              <el-rate v-model="commentForm.rating" />
            </el-form-item>
            <el-form-item label="内容">
              <el-input v-model="commentForm.content" type="textarea" :rows="2" />
            </el-form-item>
            <el-button type="primary" @click="submitComment">发表评论</el-button>
          </el-form>
          <CommentList :activity-id="activity.id" />
        </el-tab-pane>
      </el-tabs>

      <el-button class="fav" :type="favorited ? 'warning' : 'default'" @click="toggleFavorite">
        {{ favorited ? '已收藏' : '收藏' }}
      </el-button>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import SignupForm from '@/components/common/SignupForm.vue'
import CommentList from '@/components/common/CommentList.vue'
import { useActivityStore } from '@/stores/activityStore'
import { useCommentStore } from '@/stores/commentStore'
import { useAuth } from '@/hooks/useAuth'
import { createComment } from '@/api/comment'
import { addFavorite, removeFavorite, listMyFavorites } from '@/api/favorite'
import { ActivityStatusText, ActivityTypeText } from '@/constants/activity'
import { formatDateTime } from '@/utils/dateFormat'
import type { Activity } from '@/types'

const route = useRoute()
const store = useActivityStore()
const commentStore = useCommentStore()
const auth = useAuth()
const loading = ref(false)
const activity = ref<Activity | null>(null)
const registeredCount = ref(0)
const favorited = ref(false)
const commentForm = reactive({ rating: 5, content: '' })

const canSignup = computed(() => activity.value?.status === 'published')

async function checkFavorited(id: number) {
  const res = await listMyFavorites({ page: 1, page_size: 200 })
  return res.data.list.some((f: { activity_id: number }) => f.activity_id === id)
}

async function load() {
  loading.value = true
  try {
    const id = Number(route.params.id)
    activity.value = await store.fetchDetail(id)
    registeredCount.value = store.registeredCount
    if (auth.isLoggedIn) {
      favorited.value = await checkFavorited(id)
    }
  } finally {
    loading.value = false
  }
}

async function submitComment() {
  if (!activity.value) return
  await createComment(activity.value.id, commentForm)
  ElMessage.success('评论成功')
  commentForm.content = ''
  await commentStore.fetchList(activity.value.id)
}

async function toggleFavorite() {
  if (!activity.value) return
  if (favorited.value) {
    await removeFavorite(activity.value.id)
    favorited.value = false
  } else {
    await addFavorite(activity.value.id)
    favorited.value = true
  }
}

function onSignup() {
  ElMessage.success('报名成功，可在个人中心查看')
}

onMounted(load)
</script>

<style scoped>
.meta { display: flex; gap: 16px; align-items: center; color: #606266; flex-wrap: wrap; }
.desc { color: #303133; line-height: 1.7; }
.mb-2 { margin-bottom: 12px; }
.fav { margin-top: 16px; }
</style>
