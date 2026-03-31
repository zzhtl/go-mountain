import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '../api'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref(JSON.parse(localStorage.getItem('userInfo') || 'null'))

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => userInfo.value?.role === 'admin')
  const username = computed(() => userInfo.value?.username || '')
  const roleDisplay = computed(() => userInfo.value?.role_display || userInfo.value?.role || '用户')

  const login = async (loginForm) => {
    const data = await authApi.login(loginForm)
    token.value = data.token
    userInfo.value = data.user
    localStorage.setItem('token', data.token)
    localStorage.setItem('userInfo', JSON.stringify(data.user))
  }

  const logout = () => {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('userInfo')
  }

  const changePassword = async (data) => {
    await authApi.changePassword(data)
  }

  return {
    token,
    userInfo,
    isLoggedIn,
    isAdmin,
    username,
    roleDisplay,
    login,
    logout,
    changePassword
  }
})
