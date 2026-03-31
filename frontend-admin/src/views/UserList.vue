<template>
  <div class="user-list">
    <h1>用户列表</h1>
    <table>
      <thead>
        <tr>
          <th>ID</th>
          <th>手机号</th>
          <th>OpenID</th>
          <th>姓名</th>
          <th>创建时间</th>
          <th>更新时间</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in users" :key="item.id">
          <td><router-link :to="`/admin/users/${item.id}`">{{ item.id }}</router-link></td>
          <td>{{ item.phone || '-' }}</td>
          <td>{{ item.openid || '-' }}</td>
          <td>{{ item.name || '-' }}</td>
          <td>{{ item.created_at }}</td>
          <td>{{ item.updated_at }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { userApi } from '../api'

const users = ref([])
const error = ref('')

onMounted(async () => {
  try {
    const data = await userApi.list()
    users.value = data.list || data || []
  } catch (e) {
    error.value = '加载用户失败'
  }
})
</script>

<style scoped>
.user-list { padding: 1em; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 0.5em; border: 1px solid #ccc; text-align: left; }
</style> 