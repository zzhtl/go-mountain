<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>用户详情</span>
        <el-button @click="$router.back()">返回</el-button>
      </div>
    </template>
    
    <el-descriptions v-if="user" :column="1" border>
      <el-descriptions-item label="ID">{{ user.id }}</el-descriptions-item>
      <el-descriptions-item label="手机号">{{ user.phone || '-' }}</el-descriptions-item>
      <el-descriptions-item label="OpenID">{{ user.openid || '-' }}</el-descriptions-item>
      <el-descriptions-item label="姓名">{{ user.name || '-' }}</el-descriptions-item>
      <el-descriptions-item label="创建时间">{{ user.created_at }}</el-descriptions-item>
      <el-descriptions-item label="更新时间">{{ user.updated_at }}</el-descriptions-item>
    </el-descriptions>
    
    <div v-else class="loading">
      加载中...
    </div>
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import axios from 'axios'

const route = useRoute()
const user = ref(null)

onMounted(async () => {
  try {
    const res = await axios.get(`/api/admin/users/${route.params.id}`)
    user.value = res.data
  } catch (error) {
    console.error('加载用户失败', error)
  }
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.loading {
  text-align: center;
  padding: 50px;
  color: #999;
}
</style> 