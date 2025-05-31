import { createRouter, createWebHistory } from 'vue-router'
import Login from '../views/Login.vue'
import Layout from '../views/Layout.vue'
import UserList from '../views/UserList.vue'
import UserDetail from '../views/UserDetail.vue'
import ColumnList from '../views/ColumnList.vue'
import ArticleList from '../views/ArticleList.vue'
import ArticleEdit from '../views/ArticleEdit.vue'
import BackendUserList from '../views/BackendUserList.vue'
import ChangePassword from '../views/ChangePassword.vue'
import RoleList from '../views/RoleList.vue'
import MenuList from '../views/MenuList.vue'

const routes = [
  { 
    path: '/', 
    redirect: '/login'
  },
  { 
    path: '/login', 
    component: Login,
    meta: { requiresAuth: false }
  },
  {
    path: '/admin',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: 'users',
        component: UserList
      },
      {
        path: 'users/:id',
        component: UserDetail
      },
      {
        path: 'columns',
        component: ColumnList
      },
      {
        path: 'articles',
        component: ArticleList
      },
      {
        path: 'articles/create',
        component: ArticleEdit
      },
      {
        path: 'articles/edit/:id',
        component: ArticleEdit
      },
      {
        path: 'backend-users',
        component: BackendUserList
      },
      {
        path: 'roles',
        component: RoleList
      },
      {
        path: 'menus',
        component: MenuList
      },
      {
        path: 'change-password',
        component: ChangePassword
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

router.beforeEach((to, from, next) => {
  console.log('路由跳转:', to.path, '来自:', from.path)
  const token = localStorage.getItem('token')
  
  // 如果要去登录页面，且已经登录了，重定向到管理页面
  if (to.path === '/login' && token) {
    console.log('已登录，跳转到管理页面')
    return next('/admin/articles')
  }
  
  // 如果需要认证的页面，但没有token，跳转到登录页
  if (to.meta.requiresAuth && !token) {
    console.log('未登录，跳转到登录页面')
    return next('/login')
  }
  
  console.log('正常路由导航到:', to.path)
  next()
})

export default router 