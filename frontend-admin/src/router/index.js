import { createRouter, createWebHistory } from 'vue-router'
import Login from '../views/Login.vue'
import UserList from '../views/UserList.vue'
import UserDetail from '../views/UserDetail.vue'

const routes = [
  { path: '/login', component: Login },
  {
    path: '/admin/users',
    component: UserList,
    meta: { requiresAuth: true }
  },
  {
    path: '/admin/users/:id',
    component: UserDetail,
    meta: { requiresAuth: true }
  },
  { path: '/', redirect: '/admin/users' }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  if (to.meta.requiresAuth && !localStorage.getItem('token')) {
    return next('/login')
  }
  next()
})

export default router 