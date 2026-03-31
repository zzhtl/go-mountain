import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    redirect: '/login'
  },
  {
    path: '/login',
    component: () => import('../views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/admin',
    name: 'admin',
    component: () => import('../views/Layout.vue'),
    meta: { requiresAuth: true },
    children: [] // 动态路由在登录后注入
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

// 标记是否已添加动态路由
let dynamicRoutesAdded = false

export const resetDynamicRoutes = () => {
  dynamicRoutesAdded = false
}

router.beforeEach(async (to, from, next) => {
  const token = localStorage.getItem('token')

  // 去登录页且已登录 → 跳管理页
  if (to.path === '/login' && token) {
    return next('/admin/articles')
  }

  // 需要认证但没 token → 跳登录
  if (to.meta.requiresAuth && !token) {
    return next('/login')
  }

  // 已登录但还没加载动态路由
  if (token && !dynamicRoutesAdded) {
    dynamicRoutesAdded = true
    try {
      // 延迟导入避免循环依赖
      const { usePermissionStore } = await import('../store/permission')
      const permissionStore = usePermissionStore()
      const routes = await permissionStore.loadMenus()

      // 将动态路由添加到 /admin 下
      for (const route of routes) {
        router.addRoute('admin', route)
      }

      // 给 /admin 路由加 name 以便 addRoute 使用
      // 重新导航到目标路由（因为此时路由表已更新）
      return next({ ...to, replace: true })
    } catch (error) {
      console.error('加载动态路由失败:', error)
      localStorage.removeItem('token')
      localStorage.removeItem('userInfo')
      dynamicRoutesAdded = false
      return next('/login')
    }
  }

  next()
})

export default router
