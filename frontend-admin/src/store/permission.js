import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { backendUserApi } from '../api'

// 前端页面组件映射（静态已有的页面）
const componentMap = {
  'articles': () => import('../views/ArticleList.vue'),
  'articles/create': () => import('../views/ArticleEdit.vue'),
  'articles/edit/:id': () => import('../views/ArticleEdit.vue'),
  'columns': () => import('../views/ColumnList.vue'),
  'users': () => import('../views/UserList.vue'),
  'users/:id': () => import('../views/UserDetail.vue'),
  'backend-users': () => import('../views/BackendUserList.vue'),
  'roles': () => import('../views/RoleList.vue'),
  'menus': () => import('../views/MenuList.vue'),
  'change-password': () => import('../views/ChangePassword.vue'),
  'system-configs': () => import('../views/system/SystemConfig.vue'),
  'activities': () => import('../views/business/ActivityList.vue'),
  'activities/create': () => import('../views/business/ActivityEdit.vue'),
  'activities/edit/:id': () => import('../views/business/ActivityEdit.vue'),
  'registrations': () => import('../views/business/RegistrationList.vue'),
  'payments': () => import('../views/business/PaymentList.vue'),
}

export const usePermissionStore = defineStore('permission', () => {
  const menus = ref([])
  const permissions = ref([])
  const dynamicRoutes = ref([])
  const loaded = ref(false)

  // 从菜单树中提取所有权限标识（type=3 的按钮权限）
  const extractPermissions = (menuTree) => {
    const perms = []
    const walk = (items) => {
      for (const item of items) {
        if (item.type === 3 && item.permission) {
          perms.push(item.permission)
        }
        if (item.children?.length) {
          walk(item.children)
        }
      }
    }
    walk(menuTree)
    return perms
  }

  // 从菜单树生成动态路由
  const generateRoutes = (menuTree) => {
    const routes = []
    const walk = (items) => {
      for (const item of items) {
        // type=2 是菜单页面，有 path 才生成路由
        if (item.type === 2 && item.path) {
          const routePath = item.path.replace(/^\/admin\//, '')
          const component = componentMap[routePath]
          if (component) {
            routes.push({
              path: routePath,
              component,
              meta: { title: item.name, icon: item.icon }
            })
          }
        }
        if (item.children?.length) {
          walk(item.children)
        }
      }
    }
    walk(menuTree)

    // 始终添加修改密码路由和文章编辑子路由
    const extras = [
      { path: 'change-password', component: componentMap['change-password'], meta: { title: '修改密码', hidden: true } },
      { path: 'articles/create', component: componentMap['articles/create'], meta: { title: '创建文章', hidden: true } },
      { path: 'articles/edit/:id', component: componentMap['articles/edit/:id'], meta: { title: '编辑文章', hidden: true } },
      { path: 'users/:id', component: componentMap['users/:id'], meta: { title: '用户详情', hidden: true } },
      { path: 'activities/create', component: componentMap['activities/create'], meta: { title: '创建活动', hidden: true } },
      { path: 'activities/edit/:id', component: componentMap['activities/edit/:id'], meta: { title: '编辑活动', hidden: true } },
    ]
    for (const extra of extras) {
      if (extra.component && !routes.find(r => r.path === extra.path)) {
        routes.push(extra)
      }
    }

    return routes
  }

  // 从后端加载用户菜单并生成路由
  const loadMenus = async () => {
    try {
      const data = await backendUserApi.currentMenus()
      menus.value = data || []
      permissions.value = extractPermissions(menus.value)
      dynamicRoutes.value = generateRoutes(menus.value)
      loaded.value = true
      return dynamicRoutes.value
    } catch (error) {
      console.error('加载菜单失败:', error)
      menus.value = []
      permissions.value = []
      dynamicRoutes.value = []
      loaded.value = true
      return []
    }
  }

  // 检查是否拥有指定权限
  const hasPermission = (perm) => {
    // admin 拥有所有权限（由后端保证，前端也做一层）
    const userInfo = JSON.parse(localStorage.getItem('userInfo') || 'null')
    if (userInfo?.role === 'admin') return true
    return permissions.value.includes(perm)
  }

  const reset = () => {
    menus.value = []
    permissions.value = []
    dynamicRoutes.value = []
    loaded.value = false
  }

  return {
    menus,
    permissions,
    dynamicRoutes,
    loaded,
    loadMenus,
    hasPermission,
    reset
  }
})
