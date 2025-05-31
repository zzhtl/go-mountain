<template>
  <div class="user-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>用户管理</span>
          <el-button 
            type="primary" 
            @click="showCreateDialog = true"
            v-if="canManageUsers"
          >
            <el-icon><Plus /></el-icon>
            新增用户
          </el-button>
        </div>
      </template>

      <el-table :data="users" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" width="150" />
        <el-table-column prop="email" label="邮箱" width="200" />
        <el-table-column prop="role" label="角色" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.role === 'admin' ? 'danger' : 'primary'">
              {{ scope.row.role === 'admin' ? '管理员' : '编辑员' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : 'danger'">
              {{ scope.row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300" v-if="canManageUsers">
          <template #default="scope">
            <el-button size="small" @click="editUser(scope.row)">编辑</el-button>
            <el-button 
              size="small" 
              :type="scope.row.status === 1 ? 'warning' : 'success'"
              @click="toggleUserStatus(scope.row)"
            >
              {{ scope.row.status === 1 ? '禁用' : '启用' }}
            </el-button>
            <el-button 
              size="small" 
              type="danger"
              @click="resetPassword(scope.row)"
            >
              重置密码
            </el-button>
            <el-button 
              size="small" 
              type="danger" 
              @click="deleteUser(scope.row)"
              :disabled="scope.row.id === currentUserId"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="pagination"
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="fetchUsers"
        @size-change="fetchUsers"
      />
    </el-card>

    <!-- 创建用户对话框 -->
    <el-dialog v-model="showCreateDialog" title="新增用户" width="500px">
      <el-form :model="createForm" :rules="createRules" ref="createFormRef" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="createForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="createForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="createForm.role" placeholder="请选择角色">
            <el-option label="管理员" value="admin" />
            <el-option label="编辑员" value="editor" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="createUser" :loading="createLoading">确定</el-button>
      </template>
    </el-dialog>

    <!-- 编辑用户对话框 -->
    <el-dialog v-model="showEditDialog" title="编辑用户" width="500px">
      <el-form :model="editForm" :rules="createRules" ref="editFormRef" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="editForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="editForm.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="editForm.role" placeholder="请选择角色">
            <el-option label="管理员" value="admin" />
            <el-option label="编辑员" value="editor" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="updateUser" :loading="editLoading">确定</el-button>
      </template>
    </el-dialog>

    <!-- 重置密码对话框 -->
    <el-dialog v-model="showPasswordDialog" title="重置密码" width="400px">
      <p>确定要重置用户 <strong>{{ selectedUser?.username }}</strong> 的密码吗？</p>
      <p class="warning-text">重置后的新密码为：</p>
      <el-input 
        v-if="newPassword" 
        v-model="newPassword" 
        readonly 
        class="password-display"
      >
        <template #append>
          <el-button @click="copyPassword">复制</el-button>
        </template>
      </el-input>
      <template #footer>
        <el-button @click="showPasswordDialog = false">取消</el-button>
        <el-button 
          type="primary" 
          @click="confirmResetPassword" 
          :loading="resetLoading"
          v-if="!newPassword"
        >
          确定重置
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import axios from 'axios'

const loading = ref(false)
const createLoading = ref(false)
const editLoading = ref(false)
const resetLoading = ref(false)
const users = ref([])
const showCreateDialog = ref(false)
const showEditDialog = ref(false)
const showPasswordDialog = ref(false)
const selectedUser = ref(null)
const newPassword = ref('')

const pagination = ref({
  page: 1,
  pageSize: 20,
  total: 0
})

const createForm = ref({
  username: '',
  email: '',
  role: 'editor'
})

const editForm = ref({
  id: null,
  username: '',
  email: '',
  role: 'editor'
})

const createFormRef = ref()
const editFormRef = ref()

const createRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
  ],
  role: [
    { required: true, message: '请选择角色', trigger: 'change' }
  ]
}

// 获取当前用户信息
const currentUserId = ref(null)
const currentUserRole = ref('')

// 检查是否有管理用户的权限
const canManageUsers = computed(() => {
  return currentUserRole.value === 'admin'
})

onMounted(() => {
  fetchUsers()
  getCurrentUser()
})

const getCurrentUser = () => {
  // 从localStorage获取用户信息
  const userInfo = JSON.parse(localStorage.getItem('userInfo') || '{}')
  currentUserId.value = userInfo.id
  currentUserRole.value = userInfo.role
}

const fetchUsers = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/admin/backend-users', {
      params: {
        page: pagination.value.page,
        page_size: pagination.value.pageSize
      }
    })
    users.value = response.data.list || []
    pagination.value.total = response.data.total || 0
  } catch (error) {
    ElMessage.error('获取用户列表失败')
  } finally {
    loading.value = false
  }
}

const createUser = async () => {
  if (!createFormRef.value) return
  
  const valid = await createFormRef.value.validate().catch(() => false)
  if (!valid) return

  createLoading.value = true
  try {
    const response = await axios.post('/api/admin/backend-users', createForm.value)
    ElMessage.success('用户创建成功')
    showCreateDialog.value = false
    
    // 显示生成的密码
    ElMessageBox.alert(
      `用户创建成功！\n初始密码：${response.data.password}\n请及时告知用户并要求其修改密码。`,
      '创建成功',
      {
        confirmButtonText: '确定',
        type: 'success'
      }
    )
    
    // 重置表单
    createForm.value = {
      username: '',
      email: '',
      role: 'editor'
    }
    fetchUsers()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '创建用户失败')
  } finally {
    createLoading.value = false
  }
}

const editUser = (user) => {
  editForm.value = {
    id: user.id,
    username: user.username,
    email: user.email,
    role: user.role
  }
  showEditDialog.value = true
}

const updateUser = async () => {
  if (!editFormRef.value) return
  
  const valid = await editFormRef.value.validate().catch(() => false)
  if (!valid) return

  editLoading.value = true
  try {
    await axios.put(`/api/admin/backend-users/${editForm.value.id}`, {
      username: editForm.value.username,
      email: editForm.value.email,
      role: editForm.value.role
    })
    ElMessage.success('用户更新成功')
    showEditDialog.value = false
    fetchUsers()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '更新用户失败')
  } finally {
    editLoading.value = false
  }
}

const toggleUserStatus = async (user) => {
  const newStatus = user.status === 1 ? 0 : 1
  const action = newStatus === 1 ? '启用' : '禁用'
  
  try {
    await ElMessageBox.confirm(
      `确定要${action}用户 ${user.username} 吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    await axios.put(`/api/admin/backend-users/${user.id}/status`, {
      status: newStatus
    })
    ElMessage.success(`${action}成功`)
    fetchUsers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || `${action}失败`)
    }
  }
}

const resetPassword = (user) => {
  selectedUser.value = user
  newPassword.value = ''
  showPasswordDialog.value = true
}

const confirmResetPassword = async () => {
  resetLoading.value = true
  try {
    const response = await axios.put(`/api/admin/backend-users/${selectedUser.value.id}/reset-password`)
    newPassword.value = response.data.password
    ElMessage.success('密码重置成功')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '重置密码失败')
    showPasswordDialog.value = false
  } finally {
    resetLoading.value = false
  }
}

const copyPassword = async () => {
  try {
    await navigator.clipboard.writeText(newPassword.value)
    ElMessage.success('密码已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败，请手动复制')
  }
}

const deleteUser = async (user) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除用户 ${user.username} 吗？此操作不可逆！`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    await axios.delete(`/api/admin/backend-users/${user.id}`)
    ElMessage.success('删除成功')
    fetchUsers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '删除失败')
    }
  }
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleString('zh-CN')
}
</script>

<style scoped>
.user-management {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.pagination {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.warning-text {
  color: #e6a23c;
  font-weight: bold;
  margin: 10px 0;
}

.password-display {
  margin-top: 10px;
}

.password-display :deep(.el-input__inner) {
  font-family: monospace;
  font-weight: bold;
  color: #409eff;
}
</style> 