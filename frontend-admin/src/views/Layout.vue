<template>
  <el-container class="layout-container">
    <el-header>
      <div class="header-content">
        <h1>远山公益后台管理系统</h1>
        <div class="header-right">
          <span class="user-info" v-if="userInfo">
            欢迎，{{ userInfo.username }} ({{ userInfo.role_display || userInfo.role || '用户' }})
          </span>
          <el-dropdown @command="handleCommand">
            <el-button text>
              <el-icon><Setting /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="changePassword">修改密码</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
    </el-header>
    
    <el-container>
      <el-aside width="200px">
        <el-menu
          :default-active="activeMenu"
          router
          class="el-menu-vertical"
        >
          <el-menu-item index="/admin/articles" v-if="hasMenuAccess('articles')">
            <el-icon><Document /></el-icon>
            <span>文章管理</span>
          </el-menu-item>
          
          <el-menu-item index="/admin/columns" v-if="hasMenuAccess('columns')">
            <el-icon><Menu /></el-icon>
            <span>栏目管理</span>
          </el-menu-item>
          
          <el-menu-item index="/admin/users" v-if="hasMenuAccess('mp-users')">
            <el-icon><User /></el-icon>
            <span>小程序用户</span>
          </el-menu-item>
          
          <el-menu-item index="/admin/backend-users" v-if="hasMenuAccess('backend-users')">
            <el-icon><UserFilled /></el-icon>
            <span>用户管理</span>
          </el-menu-item>
          
          <el-menu-item index="/admin/roles" v-if="hasMenuAccess('roles')">
            <el-icon><Key /></el-icon>
            <span>角色管理</span>
          </el-menu-item>
          
          <el-menu-item index="/admin/menus" v-if="hasMenuAccess('menus')">
            <el-icon><Grid /></el-icon>
            <span>菜单管理</span>
          </el-menu-item>
        </el-menu>
      </el-aside>
      
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Setting, Document, Menu, User, UserFilled, Key, Grid } from '@element-plus/icons-vue'
import axios from 'axios'

const router = useRouter()
const route = useRoute()
const userInfo = ref(null)
const userMenus = ref([])

const activeMenu = computed(() => {
  const path = route.path
  if (path.startsWith('/admin/articles')) return '/admin/articles'
  if (path.startsWith('/admin/columns')) return '/admin/columns'
  if (path.startsWith('/admin/users') && !path.includes('backend')) return '/admin/users'
  if (path.startsWith('/admin/backend-users')) return '/admin/backend-users'
  if (path.startsWith('/admin/roles')) return '/admin/roles'
  if (path.startsWith('/admin/menus')) return '/admin/menus'
  return path
})

const isAdmin = computed(() => {
  return userInfo.value?.role === 'admin'
})

// 检查用户是否有指定菜单的访问权限
const hasMenuAccess = (menuName) => {
  if (!userMenus.value || userMenus.value.length === 0) {
    // 如果菜单权限还没加载，先按角色判断（兼容性）
    return isAdmin.value
  }
  
  // 检查用户菜单中是否包含指定菜单
  const hasAccess = userMenus.value.some(menu => 
    menu.name === menuName || 
    (menu.children && menu.children.some(child => child.name === menuName))
  )
  return hasAccess
}

onMounted(() => {
  // 获取用户信息
  const savedUserInfo = localStorage.getItem('userInfo')
  if (savedUserInfo) {
    userInfo.value = JSON.parse(savedUserInfo)
  }
  
  // 获取用户菜单权限
  fetchUserMenus()
})

const fetchUserMenus = async () => {
  try {
    const response = await axios.get('/api/admin/backend-users/current/menus')
    userMenus.value = response.data || []
  } catch (error) {
    console.error('获取用户菜单权限失败:', error)
    // 如果获取失败，保持原有的角色判断逻辑
  }
}

const handleCommand = (command) => {
  if (command === 'logout') {
    logout()
  } else if (command === 'changePassword') {
    router.push('/admin/change-password')
  }
}

const logout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('userInfo')
  delete axios.defaults.headers.common['Authorization']
  ElMessage.success('退出成功')
  router.push('/login')
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.el-header {
  background-color: #545c64;
  color: white;
  line-height: 60px;
  height: 60px;
  padding: 0 20px;
  flex-shrink: 0;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-content h1 {
  margin: 0;
  font-size: 20px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 15px;
}

.user-info {
  color: #ccc;
  font-size: 14px;
}

.el-container {
  flex: 1;
  overflow: hidden;
}

.el-aside {
  background-color: #f5f5f5;
  height: calc(100vh - 60px);
  overflow-y: auto;
}

.el-menu-vertical {
  height: 100%;
  border-right: none;
}

.el-main {
  background-color: #f0f2f5;
  padding: 20px;
  height: calc(100vh - 60px);
  overflow-y: auto;
}
</style> 