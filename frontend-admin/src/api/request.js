import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 30000
})

// 请求拦截器
request.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => Promise.reject(error)
)

// 响应拦截器：统一处理新的响应格式 { code, message, data }
request.interceptors.response.use(
  response => {
    const res = response.data
    // 如果有 code 字段，按新格式处理
    if (res.code !== undefined) {
      if (res.code === 0) {
        return res.data
      }
      ElMessage.error(res.message || '请求失败')
      return Promise.reject(new Error(res.message))
    }
    // 兼容旧格式（直接返回 data）
    return res
  },
  error => {
    const status = error.response?.status
    if (status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('userInfo')
      router.push('/login')
      ElMessage.error('登录已过期，请重新登录')
    } else if (status === 403) {
      ElMessage.error('无权限访问')
    } else {
      const msg = error.response?.data?.message || error.response?.data?.error || '请求失败'
      ElMessage.error(msg)
    }
    return Promise.reject(error)
  }
)

export default request
