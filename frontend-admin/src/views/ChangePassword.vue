<template>
  <div class="change-password">
    <el-card style="max-width: 500px; margin: 20px auto;">
      <template #header>
        <h3>修改密码</h3>
      </template>
      
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="原密码" prop="oldPassword">
          <el-input 
            v-model="form.oldPassword" 
            type="password" 
            placeholder="请输入原密码"
            show-password
          />
        </el-form-item>
        
        <el-form-item label="新密码" prop="newPassword">
          <el-input 
            v-model="form.newPassword" 
            type="password" 
            placeholder="请输入新密码（至少6位）"
            show-password
          />
        </el-form-item>
        
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input 
            v-model="form.confirmPassword" 
            type="password" 
            placeholder="请再次输入新密码"
            show-password
          />
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="changePassword" :loading="loading">
            修改密码
          </el-button>
          <el-button @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { useUserStore } from '../store/user'
import { usePermissionStore } from '../store/permission'
import { resetDynamicRoutes } from '../router'

const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)
const formRef = ref()

const form = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const validateConfirmPassword = (rule, value, callback) => {
  if (value !== form.value.newPassword) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules = {
  oldPassword: [
    { required: true, message: '请输入原密码', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

const changePassword = async () => {
  if (!formRef.value) return
  
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await userStore.changePassword({
      old_password: form.value.oldPassword,
      new_password: form.value.newPassword
    })

    ElMessage.success('密码修改成功，请重新登录')

    // 清除登录状态
    userStore.logout()
    usePermissionStore().reset()
    resetDynamicRoutes()
    router.push('/login')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '修改密码失败')
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  form.value = {
    oldPassword: '',
    newPassword: '',
    confirmPassword: ''
  }
  if (formRef.value) {
    formRef.value.clearValidate()
  }
}
</script>

<style scoped>
.change-password {
  padding: 20px;
}
</style> 