<template>
  <div>
    <el-upload
      :show-file-list="false"
      :http-request="upload"
      accept="image/*"
    >
      <img v-if="modelValue" :src="modelValue" class="preview" alt="封面" />
      <el-icon v-else class="placeholder"><Plus /></el-icon>
    </el-upload>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus'
import request from '@/utils/request'

const props = defineProps<{ modelValue: string }>()
const emit = defineEmits<{ (e: 'update:modelValue', value: string): void }>()

async function upload(options: { file: File }) {
  const form = new FormData()
  form.append('file', options.file)
  try {
    const res = await request.post('/upload/image', form, { headers: { 'Content-Type': 'multipart/form-data' } })
    emit('update:modelValue', res.data.url)
    ElMessage.success('上传成功')
  } catch {
    ElMessage.error('上传失败')
  }
}
</script>

<style scoped>
.preview { width: 160px; height: 100px; object-fit: cover; border-radius: 6px; }
.placeholder { width: 160px; height: 100px; border: 1px dashed #d9d9d9; border-radius: 6px; display: flex; align-items: center; justify-content: center; color: #909399; }
</style>
