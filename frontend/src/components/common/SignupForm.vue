<template>
  <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
    <el-form-item label="姓名" prop="name">
      <el-input v-model="form.name" placeholder="请输入姓名" />
    </el-form-item>
    <el-form-item label="手机号" prop="phone">
      <el-input v-model="form.phone" placeholder="请输入手机号" />
    </el-form-item>
    <el-form-item label="备注">
      <el-input v-model="form.remark" type="textarea" :rows="2" placeholder="备注信息（选填）" />
    </el-form-item>
    <el-form-item>
      <el-button type="primary" :loading="loading" @click="submit">立即报名</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { signup } from '@/api/registration'

const props = defineProps<{ activityId: number }>()
const emit = defineEmits<{ (e: 'success'): void }>()

const formRef = ref<FormInstance>()
const loading = ref(false)
const form = reactive({ name: '', phone: '', remark: '' })
const rules: FormRules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  phone: [{ required: true, message: '请输入手机号', trigger: 'blur' }],
}

async function submit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  loading.value = true
  try {
    await signup({ activity_id: props.activityId, ...form })
    ElMessage.success('报名成功')
    emit('success')
  } finally {
    loading.value = false
  }
}
</script>
