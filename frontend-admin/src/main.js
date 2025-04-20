import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import axios from 'axios'
import './style.css'

// 设置 API 基础路径
axios.defaults.baseURL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
// 设置可选的 Authorization 头
const token = localStorage.getItem('token')
if (token) {
  axios.defaults.headers.common['Authorization'] = `Bearer ${token}`
}

const app = createApp(App)
app.config.globalProperties.$axios = axios
app.use(router)
app.mount('#app')
