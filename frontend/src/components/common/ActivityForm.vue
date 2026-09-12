<template>
  <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
    <el-form-item label="活动标题" prop="title">
      <el-input v-model="form.title" />
    </el-form-item>
    <el-form-item label="活动类型" prop="activity_type">
      <el-select v-model="form.activity_type" style="width: 100%">
        <el-option v-for="opt in ActivityTypeOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
      </el-select>
    </el-form-item>
    <el-form-item label="活动描述">
      <el-input v-model="form.description" type="textarea" :rows="3" />
    </el-form-item>
    <el-form-item label="封面图">
      <ImageUploader v-model="form.cover_image" />
    </el-form-item>
    <el-form-item label="开始时间" prop="start_time">
      <el-date-picker v-model="form.start_time" type="datetime" style="width: 100%" />
    </el-form-item>
    <el-form-item label="结束时间" prop="end_time">
      <el-date-picker v-model="form.end_time" type="datetime" style="width: 100%" />
    </el-form-item>
    <el-form-item label="报名截止" prop="signup_deadline">
      <el-date-picker v-model="form.signup_deadline" type="datetime" style="width: 100%" />
    </el-form-item>
    <el-form-item label="地点">
      <el-input v-model="form.location" />
    </el-form-item>
    <el-form-item label="名额">
      <el-input-number v-model="form.capacity" :min="0" />
    </el-form-item>
    <el-form-item>
      <el-button type="primary" :loading="loading" @click="submit">{{ form.id ? '保存' : '创建' }}</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import ImageUploader from '@/components/common/ImageUploader.vue'
import { createActivity, updateActivity } from '@/api/activity'
import { ActivityTypeOptions } from '@/constants/activity'

const props = defineProps<{ activity?: any }>()
const emit = defineEmits<{ (e: 'success'): void }>()

const formRef = ref<FormInstance>()
const loading = ref(false)
const form = reactive({
  id: props.activity?.id || 0,
  title: props.activity?.title || '',
  description: props.activity?.description || '',
  cover_image: props.activity?.cover_image || '',
  activity_type: props.activity?.activity_type || 'lecture',
  start_time: props.activity?.start_time || '',
  end_time: props.activity?.end_time || '',
  signup_deadline: props.activity?.signup_deadline || '',
  location: props.activity?.location || '',
  capacity: props.activity?.capacity ?? 100,
})

const rules: FormRules = {
  title: [{ required: true, message: '请输入活动标题', trigger: 'blur' }],
  activity_type: [{ required: true, message: '请选择活动类型', trigger: 'change' }],
  start_time: [{ required: true, message: '请选择开始时间', trigger: 'change' }],
  end_time: [{ required: true, message: '请选择结束时间', trigger: 'change' }],
  signup_deadline: [{ required: true, message: '请选择报名截止时间', trigger: 'change' }],
}

async function submit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    if (form.id) {
      await updateActivity(form.id, { ...form })
    } else {
      await createActivity({ ...form })
    }
    ElMessage.success('保存成功')
    emit('success')
  } finally {
    loading.value = false
  }
}
</script>
