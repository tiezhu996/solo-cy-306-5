<template>
  <div>
    <el-tabs v-model="tab">
      <el-tab-pane label="基本资料" name="info">
        <el-form :model="profile" label-width="80px" class="form">
          <el-form-item label="用户名">
            <el-input :model-value="auth.auth.user?.username" disabled />
          </el-form-item>
          <el-form-item label="昵称">
            <el-input v-model="profile.nickname" />
          </el-form-item>
          <el-form-item label="邮箱">
            <el-input v-model="profile.email" />
          </el-form-item>
          <el-form-item label="手机">
            <el-input v-model="profile.phone" />
          </el-form-item>
          <el-button type="primary" @click="save">保存</el-button>
        </el-form>
      </el-tab-pane>
      <el-tab-pane label="我的报名" name="regs">
        <MyRegistrations />
      </el-tab-pane>
      <el-tab-pane label="我的收藏" name="favs">
        <FavoriteList />
      </el-tab-pane>
      <el-tab-pane label="我的评论" name="comments">
        <el-table :data="comments" border>
          <el-table-column prop="id" label="ID" width="70" />
          <el-table-column prop="activity_id" label="活动ID" width="90" />
          <el-table-column prop="rating" label="评分" width="90" />
          <el-table-column prop="content" label="内容" />
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="消息通知" name="notifs">
        <NotificationList />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import MyRegistrations from '@/components/common/MyRegistrations.vue'
import FavoriteList from '@/components/common/FavoriteList.vue'
import NotificationList from '@/components/common/NotificationList.vue'
import { useAuth } from '@/hooks/useAuth'
import { useUserStore } from '@/stores/userStore'
import { listMyComments } from '@/api/comment'
import type { CommentItem } from '@/types'

const auth = useAuth()
const userStore = useUserStore()
const tab = ref('info')
const profile = reactive({ nickname: '', email: '', phone: '' })
const comments = ref<CommentItem[]>([])

async function save() {
  const user = await userStore.updateProfile({ ...profile })
  auth.auth.user = user
  ElMessage.success('保存成功')
}

onMounted(async () => {
  const me = await userStore.fetchMe()
  profile.nickname = me.nickname || ''
  profile.email = me.email || ''
  profile.phone = me.phone || ''
  const res = await listMyComments()
  comments.value = res.data
})
</script>

<style scoped>
.form { max-width: 480px; }
</style>
