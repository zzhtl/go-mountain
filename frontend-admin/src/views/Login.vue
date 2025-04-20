<template>
  <div class="login">
    <h1>Admin Login</h1>
    <form @submit.prevent="onSubmit">
      <div>
        <label>Username: <input v-model="username" required /></label>
      </div>
      <div>
        <label>Password: <input type="password" v-model="password" required /></label>
      </div>
      <div>
        <button type="submit">Login</button>
      </div>
      <p v-if="error" class="error">{{ error }}</p>
    </form>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'

const username = ref('')
const password = ref('')
const error = ref('')
const router = useRouter()

const onSubmit = async () => {
  error.value = ''
  try {
    const res = await axios.post('/admin/login', { username: username.value, password: password.value })
    const token = res.data.token
    localStorage.setItem('token', token)
    axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
    router.push('/admin/users')
  } catch (e) {
    error.value = e.response?.data?.error || 'Login failed'
  }
}
</script>

<style scoped>
.login { max-width: 300px; margin: 2em auto; }
.login div { margin-bottom: 1em; }
.error { color: red; }
</style> 