<template>
  <el-container class="layout-container">
    <el-header>
      <div class="header-content">
        <h1>远山公益后台管理系统</h1>
        <div class="header-right">
          <span class="user-info" v-if="userStore.userInfo">
            欢迎，{{ userStore.username }} ({{ userStore.roleDisplay }})
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
          <template v-for="menu in sidebarMenus" :key="menu.id">
            <!-- 有子菜单的目录 -->
            <el-sub-menu v-if="menu.children?.length" :index="'dir-' + menu.id">
              <template #title>
                <el-icon><component :is="menu.icon || 'Folder'" /></el-icon>
                <span>{{ menu.name }}</span>
              </template>
              <el-menu-item
                v-for="child in menu.children.filter(c => c.type === 2)"
                :key="child.id"
                :index="child.path"
              >
                <el-icon><component :is="child.icon || 'Document'" /></el-icon>
                <span>{{ child.name }}</span>
              </el-menu-item>
            </el-sub-menu>

            <!-- 没有子菜单的直接菜单项 (type=2) -->
            <el-menu-item v-else-if="menu.type === 2" :index="menu.path">
              <el-icon><component :is="menu.icon || 'Document'" /></el-icon>
              <span>{{ menu.name }}</span>
            </el-menu-item>
          </template>
        </el-menu>
      </el-aside>

      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Setting } from '@element-plus/icons-vue'
import { useUserStore } from '../store/user'
import { usePermissionStore } from '../store/permission'
import { resetDynamicRoutes } from '../router'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const permissionStore = usePermissionStore()

// 侧边栏只显示目录(type=1)和菜单(type=2)，不显示按钮权限(type=3)
const sidebarMenus = computed(() => {
  return permissionStore.menus.filter(m => m.type === 1 || m.type === 2)
})

const activeMenu = computed(() => {
  return route.path
})

const handleCommand = (command) => {
  if (command === 'logout') {
    logout()
  } else if (command === 'changePassword') {
    router.push('/admin/change-password')
  }
}

const logout = () => {
  userStore.logout()
  permissionStore.reset()
  resetDynamicRoutes()
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
