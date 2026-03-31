import request from './request'

// ==================== 认证 ====================
export const authApi = {
  login: data => request.post('/api/admin/backend-auth/login', data),
  changePassword: data => request.put('/api/admin/backend-auth/change-password', data)
}

// ==================== 后台用户 ====================
export const backendUserApi = {
  list: params => request.get('/api/admin/backend-users/', { params }),
  get: id => request.get(`/api/admin/backend-users/${id}`),
  create: data => request.post('/api/admin/backend-users/', data),
  update: (id, data) => request.put(`/api/admin/backend-users/${id}`, data),
  delete: id => request.delete(`/api/admin/backend-users/${id}`),
  updateStatus: (id, data) => request.put(`/api/admin/backend-users/${id}/status`, data),
  resetPassword: id => request.put(`/api/admin/backend-users/${id}/reset-password`),
  currentMenus: () => request.get('/api/admin/backend-users/current/menus')
}

// ==================== 文章 ====================
export const articleApi = {
  list: params => request.get('/api/admin/articles/', { params }),
  get: id => request.get(`/api/admin/articles/${id}`),
  create: data => request.post('/api/admin/articles/', data),
  update: (id, data) => request.put(`/api/admin/articles/${id}`, data),
  delete: id => request.delete(`/api/admin/articles/${id}`),
  updateStatus: (id, data) => request.put(`/api/admin/articles/${id}/status`, data)
}

// ==================== 栏目 ====================
export const columnApi = {
  list: () => request.get('/api/admin/columns/'),
  get: id => request.get(`/api/admin/columns/${id}`),
  create: data => request.post('/api/admin/columns/', data),
  update: (id, data) => request.put(`/api/admin/columns/${id}`, data),
  delete: id => request.delete(`/api/admin/columns/${id}`)
}

// ==================== 角色 ====================
export const roleApi = {
  list: params => request.get('/api/admin/roles/', { params }),
  get: id => request.get(`/api/admin/roles/${id}`),
  create: data => request.post('/api/admin/roles/', data),
  update: (id, data) => request.put(`/api/admin/roles/${id}`, data),
  delete: id => request.delete(`/api/admin/roles/${id}`),
  updateStatus: (id, data) => request.put(`/api/admin/roles/${id}/status`, data),
  getMenus: id => request.get(`/api/admin/roles/${id}/menus`),
  updateMenus: (id, data) => request.put(`/api/admin/roles/${id}/menus`, data)
}

// ==================== 菜单 ====================
export const menuApi = {
  list: () => request.get('/api/admin/menus/'),
  tree: () => request.get('/api/admin/menus/tree'),
  get: id => request.get(`/api/admin/menus/${id}`),
  create: data => request.post('/api/admin/menus/', data),
  update: (id, data) => request.put(`/api/admin/menus/${id}`, data),
  delete: id => request.delete(`/api/admin/menus/${id}`),
  updateStatus: (id, data) => request.put(`/api/admin/menus/${id}/status`, data)
}

// ==================== 活动管理 ====================
export const activityApi = {
  list: params => request.get('/api/admin/activities/', { params }),
  get: id => request.get(`/api/admin/activities/${id}`),
  create: data => request.post('/api/admin/activities/', data),
  update: (id, data) => request.put(`/api/admin/activities/${id}`, data),
  delete: id => request.delete(`/api/admin/activities/${id}`),
  updateStatus: (id, data) => request.put(`/api/admin/activities/${id}/status`, data)
}

// ==================== 报名管理 ====================
export const registrationApi = {
  list: params => request.get('/api/admin/registrations/', { params }),
  get: id => request.get(`/api/admin/registrations/${id}`)
}

// ==================== 支付管理 ====================
export const paymentApi = {
  list: params => request.get('/api/admin/payments/', { params }),
  get: id => request.get(`/api/admin/payments/${id}`),
  refund: id => request.put(`/api/admin/payments/${id}/refund`)
}

// ==================== 小程序用户 ====================
export const userApi = {
  list: params => request.get('/api/admin/users/', { params }),
  get: id => request.get(`/api/admin/users/${id}`),
  update: (id, data) => request.put(`/api/admin/users/${id}`, data),
  delete: id => request.delete(`/api/admin/users/${id}`)
}

// ==================== 系统配置 ====================
export const systemConfigApi = {
  list: params => request.get('/api/admin/system-configs/', { params }),
  groups: () => request.get('/api/admin/system-configs/groups'),
  save: data => request.post('/api/admin/system-configs/', data),
  batchSave: data => request.post('/api/admin/system-configs/batch', data),
  delete: key => request.delete('/api/admin/system-configs/', { params: { key } })
}

// ==================== 文件上传 ====================
export const uploadApi = {
  image: formData => request.post('/api/admin/upload/image', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),
  video: formData => request.post('/api/admin/upload/video', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
