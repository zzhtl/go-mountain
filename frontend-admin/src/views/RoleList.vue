<template>
  <div class="role-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>角色管理</span>
          <el-button type="primary" @click="showCreateDialog = true">
            <el-icon><Plus /></el-icon>
            新增角色
          </el-button>
        </div>
      </template>

      <el-table :data="roles" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="角色标识" width="150" />
        <el-table-column prop="display_name" label="显示名称" width="150" />
        <el-table-column prop="description" label="描述" />
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
        <el-table-column label="操作" width="300">
          <template #default="scope">
            <el-button size="small" @click="editRole(scope.row)">编辑</el-button>
            <el-button size="small" type="warning" @click="managePermissions(scope.row)">权限</el-button>
            <el-button 
              size="small" 
              :type="scope.row.status === 1 ? 'warning' : 'success'"
              @click="toggleRoleStatus(scope.row)"
            >
              {{ scope.row.status === 1 ? '禁用' : '启用' }}
            </el-button>
            <el-button 
              size="small" 
              type="danger" 
              @click="deleteRole(scope.row)"
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
        @current-change="fetchRoles"
        @size-change="fetchRoles"
      />
    </el-card>

    <!-- 创建角色对话框 -->
    <el-dialog v-model="showCreateDialog" title="新增角色" width="500px">
      <el-form :model="createForm" :rules="formRules" ref="createFormRef" label-width="100px">
        <el-form-item label="角色标识" prop="name">
          <el-input v-model="createForm.name" placeholder="请输入角色标识（英文）" />
        </el-form-item>
        <el-form-item label="显示名称" prop="display_name">
          <el-input v-model="createForm.display_name" placeholder="请输入显示名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input 
            v-model="createForm.description" 
            type="textarea" 
            placeholder="请输入角色描述"
            :rows="3"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="createRole" :loading="createLoading">确定</el-button>
      </template>
    </el-dialog>

    <!-- 编辑角色对话框 -->
    <el-dialog v-model="showEditDialog" title="编辑角色" width="500px">
      <el-form :model="editForm" :rules="formRules" ref="editFormRef" label-width="100px">
        <el-form-item label="角色标识" prop="name">
          <el-input v-model="editForm.name" placeholder="请输入角色标识（英文）" />
        </el-form-item>
        <el-form-item label="显示名称" prop="display_name">
          <el-input v-model="editForm.display_name" placeholder="请输入显示名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input 
            v-model="editForm.description" 
            type="textarea" 
            placeholder="请输入角色描述"
            :rows="3"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="updateRole" :loading="editLoading">确定</el-button>
      </template>
    </el-dialog>

    <!-- 权限管理对话框 -->
    <el-dialog v-model="showPermissionDialog" title="权限管理" width="600px">
      <div class="permission-content">
        <h4>为角色 "{{ selectedRole?.display_name }}" 分配菜单权限</h4>
        <el-tree
          ref="menuTreeRef"
          :data="menuTree"
          :props="treeProps"
          show-checkbox
          node-key="id"
          :default-checked-keys="checkedMenus"
          :check-strictly="false"
          class="menu-tree"
        />
      </div>
      <template #footer>
        <el-button @click="showPermissionDialog = false">取消</el-button>
        <el-button type="primary" @click="savePermissions" :loading="permissionLoading">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import axios from 'axios'

const loading = ref(false)
const createLoading = ref(false)
const editLoading = ref(false)
const permissionLoading = ref(false)
const roles = ref([])
const menuTree = ref([])
const showCreateDialog = ref(false)
const showEditDialog = ref(false)
const showPermissionDialog = ref(false)
const selectedRole = ref(null)
const checkedMenus = ref([])
const menuTreeRef = ref()

const pagination = ref({
  page: 1,
  pageSize: 20,
  total: 0
})

const createForm = ref({
  name: '',
  display_name: '',
  description: ''
})

const editForm = ref({
  id: null,
  name: '',
  display_name: '',
  description: ''
})

const createFormRef = ref()
const editFormRef = ref()

const formRules = {
  name: [
    { required: true, message: '请输入角色标识', trigger: 'blur' },
    { pattern: /^[a-zA-Z_][a-zA-Z0-9_]*$/, message: '角色标识只能包含字母、数字和下划线，且以字母或下划线开头', trigger: 'blur' }
  ],
  display_name: [
    { required: true, message: '请输入显示名称', trigger: 'blur' }
  ]
}

const treeProps = {
  children: 'children',
  label: 'title'
}

onMounted(() => {
  fetchRoles()
  fetchMenuTree()
})

const fetchRoles = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/admin/roles', {
      params: {
        page: pagination.value.page,
        page_size: pagination.value.pageSize
      }
    })
    roles.value = response.data.list || []
    pagination.value.total = response.data.total || 0
  } catch (error) {
    ElMessage.error('获取角色列表失败')
  } finally {
    loading.value = false
  }
}

const fetchMenuTree = async () => {
  try {
    const response = await axios.get('/api/admin/menus/tree')
    menuTree.value = response.data || []
  } catch (error) {
    ElMessage.error('获取菜单列表失败')
  }
}

const createRole = async () => {
  if (!createFormRef.value) return
  
  const valid = await createFormRef.value.validate().catch(() => false)
  if (!valid) return

  createLoading.value = true
  try {
    await axios.post('/api/admin/roles', createForm.value)
    ElMessage.success('角色创建成功')
    showCreateDialog.value = false
    
    // 重置表单
    createForm.value = {
      name: '',
      display_name: '',
      description: ''
    }
    fetchRoles()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '创建角色失败')
  } finally {
    createLoading.value = false
  }
}

const editRole = (role) => {
  editForm.value = {
    id: role.id,
    name: role.name,
    display_name: role.display_name,
    description: role.description
  }
  showEditDialog.value = true
}

const updateRole = async () => {
  if (!editFormRef.value) return
  
  const valid = await editFormRef.value.validate().catch(() => false)
  if (!valid) return

  editLoading.value = true
  try {
    await axios.put(`/api/admin/roles/${editForm.value.id}`, {
      name: editForm.value.name,
      display_name: editForm.value.display_name,
      description: editForm.value.description
    })
    ElMessage.success('角色更新成功')
    showEditDialog.value = false
    fetchRoles()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '更新角色失败')
  } finally {
    editLoading.value = false
  }
}

const toggleRoleStatus = async (role) => {
  const newStatus = role.status === 1 ? 0 : 1
  const action = newStatus === 1 ? '启用' : '禁用'
  
  try {
    await ElMessageBox.confirm(
      `确定要${action}角色 ${role.display_name} 吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    await axios.put(`/api/admin/roles/${role.id}/status`, {
      status: newStatus
    })
    ElMessage.success(`${action}成功`)
    fetchRoles()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || `${action}失败`)
    }
  }
}

const deleteRole = async (role) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除角色 ${role.display_name} 吗？此操作不可逆！`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    await axios.delete(`/api/admin/roles/${role.id}`)
    ElMessage.success('删除成功')
    fetchRoles()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '删除失败')
    }
  }
}

const managePermissions = async (role) => {
  selectedRole.value = role
  showPermissionDialog.value = true
  
  // 获取当前角色的菜单权限
  try {
    const response = await axios.get(`/api/admin/roles/${role.id}/menus`)
    checkedMenus.value = response.data.menu_ids || []
  } catch (error) {
    ElMessage.error('获取角色权限失败')
  }
}

const savePermissions = async () => {
  if (!selectedRole.value) return
  
  const checkedKeys = menuTreeRef.value.getCheckedKeys()
  const halfCheckedKeys = menuTreeRef.value.getHalfCheckedKeys()
  const allCheckedKeys = [...checkedKeys, ...halfCheckedKeys]
  
  permissionLoading.value = true
  try {
    await axios.put(`/api/admin/roles/${selectedRole.value.id}/menus`, {
      menu_ids: allCheckedKeys
    })
    ElMessage.success('权限保存成功')
    showPermissionDialog.value = false
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '保存权限失败')
  } finally {
    permissionLoading.value = false
  }
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleString('zh-CN')
}
</script>

<style scoped>
.role-management {
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

.permission-content {
  padding: 20px 0;
}

.menu-tree {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 10px;
  max-height: 400px;
  overflow-y: auto;
}

.permission-content h4 {
  margin-bottom: 15px;
  color: #303133;
}
</style> 