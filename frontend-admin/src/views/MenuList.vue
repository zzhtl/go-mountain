<template>
  <div class="menu-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>菜单管理</span>
          <el-button type="primary" @click="showCreateDialog = true">
            <el-icon><Plus /></el-icon>
            新增菜单
          </el-button>
        </div>
      </template>

      <el-table 
        :data="menus" 
        style="width: 100%" 
        v-loading="loading"
        row-key="id"
        :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="title" label="菜单名称" width="200" />
        <el-table-column prop="name" label="标识" width="150" />
        <el-table-column prop="path" label="路径" width="180" />
        <el-table-column prop="component" label="组件" width="150" />
        <el-table-column prop="icon" label="图标" width="100">
          <template #default="scope">
            <el-icon v-if="scope.row.icon">
              <component :is="scope.row.icon" />
            </el-icon>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column prop="type" label="类型" width="80">
          <template #default="scope">
            <el-tag :type="scope.row.type === 1 ? 'primary' : 'warning'">
              {{ scope.row.type === 1 ? '菜单' : '按钮' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : 'danger'">
              {{ scope.row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300">
          <template #default="scope">
            <el-button size="small" @click="editMenu(scope.row)">编辑</el-button>
            <el-button 
              size="small" 
              :type="scope.row.status === 1 ? 'warning' : 'success'"
              @click="toggleMenuStatus(scope.row)"
            >
              {{ scope.row.status === 1 ? '禁用' : '启用' }}
            </el-button>
            <el-button 
              size="small" 
              type="danger" 
              @click="deleteMenu(scope.row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建菜单对话框 -->
    <el-dialog v-model="showCreateDialog" title="新增菜单" width="600px">
      <el-form :model="createForm" :rules="formRules" ref="createFormRef" label-width="100px">
        <el-form-item label="父级菜单" prop="parent_id">
          <el-tree-select
            v-model="createForm.parent_id"
            :data="menuOptions"
            :props="{ children: 'children', label: 'title', value: 'id' }"
            placeholder="请选择父级菜单（可选）"
            clearable
            check-strictly
          />
        </el-form-item>
        <el-form-item label="菜单名称" prop="title">
          <el-input v-model="createForm.title" placeholder="请输入菜单名称" />
        </el-form-item>
        <el-form-item label="菜单标识" prop="name">
          <el-input v-model="createForm.name" placeholder="请输入菜单标识（英文）" />
        </el-form-item>
        <el-form-item label="路径" prop="path">
          <el-input v-model="createForm.path" placeholder="请输入路径" />
        </el-form-item>
        <el-form-item label="组件" prop="component">
          <el-input v-model="createForm.component" placeholder="请输入组件名" />
        </el-form-item>
        <el-form-item label="图标" prop="icon">
          <el-select v-model="createForm.icon" placeholder="请选择图标" clearable>
            <el-option 
              v-for="icon in iconOptions" 
              :key="icon.value" 
              :label="icon.label" 
              :value="icon.value"
            >
              <div class="icon-option">
                <el-icon><component :is="icon.value" /></el-icon>
                <span>{{ icon.label }}</span>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="createForm.sort" :min="0" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-radio-group v-model="createForm.type">
            <el-radio :label="1">菜单</el-radio>
            <el-radio :label="2">按钮</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="createMenu" :loading="createLoading">确定</el-button>
      </template>
    </el-dialog>

    <!-- 编辑菜单对话框 -->
    <el-dialog v-model="showEditDialog" title="编辑菜单" width="600px">
      <el-form :model="editForm" :rules="formRules" ref="editFormRef" label-width="100px">
        <el-form-item label="父级菜单" prop="parent_id">
          <el-tree-select
            v-model="editForm.parent_id"
            :data="menuOptions"
            :props="{ children: 'children', label: 'title', value: 'id' }"
            placeholder="请选择父级菜单（可选）"
            clearable
            check-strictly
          />
        </el-form-item>
        <el-form-item label="菜单名称" prop="title">
          <el-input v-model="editForm.title" placeholder="请输入菜单名称" />
        </el-form-item>
        <el-form-item label="菜单标识" prop="name">
          <el-input v-model="editForm.name" placeholder="请输入菜单标识（英文）" />
        </el-form-item>
        <el-form-item label="路径" prop="path">
          <el-input v-model="editForm.path" placeholder="请输入路径" />
        </el-form-item>
        <el-form-item label="组件" prop="component">
          <el-input v-model="editForm.component" placeholder="请输入组件名" />
        </el-form-item>
        <el-form-item label="图标" prop="icon">
          <el-select v-model="editForm.icon" placeholder="请选择图标" clearable>
            <el-option 
              v-for="icon in iconOptions" 
              :key="icon.value" 
              :label="icon.label" 
              :value="icon.value"
            >
              <div class="icon-option">
                <el-icon><component :is="icon.value" /></el-icon>
                <span>{{ icon.label }}</span>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="editForm.sort" :min="0" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-radio-group v-model="editForm.type">
            <el-radio :label="1">菜单</el-radio>
            <el-radio :label="2">按钮</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showEditDialog = false">取消</el-button>
        <el-button type="primary" @click="updateMenu" :loading="editLoading">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Plus, Document, Menu, User, UserFilled, Key, Grid,
  Setting, Management, DataAnalysis, Monitor
} from '@element-plus/icons-vue'
import axios from 'axios'

const loading = ref(false)
const createLoading = ref(false)
const editLoading = ref(false)
const menus = ref([])
const menuList = ref([])
const showCreateDialog = ref(false)
const showEditDialog = ref(false)

const createForm = ref({
  parent_id: 0,
  name: '',
  title: '',
  path: '',
  component: '',
  icon: '',
  sort: 0,
  type: 1
})

const editForm = ref({
  id: null,
  parent_id: 0,
  name: '',
  title: '',
  path: '',
  component: '',
  icon: '',
  sort: 0,
  type: 1
})

const createFormRef = ref()
const editFormRef = ref()

const formRules = {
  title: [
    { required: true, message: '请输入菜单名称', trigger: 'blur' }
  ],
  name: [
    { required: true, message: '请输入菜单标识', trigger: 'blur' },
    { pattern: /^[a-zA-Z_][a-zA-Z0-9_-]*$/, message: '菜单标识只能包含字母、数字、下划线和横线，且以字母或下划线开头', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择类型', trigger: 'change' }
  ]
}

const iconOptions = [
  { label: '文档', value: 'Document' },
  { label: '菜单', value: 'Menu' },
  { label: '用户', value: 'User' },
  { label: '用户（填充）', value: 'UserFilled' },
  { label: '钥匙', value: 'Key' },
  { label: '网格', value: 'Grid' },
  { label: '设置', value: 'Setting' },
  { label: '管理', value: 'Management' },
  { label: '数据分析', value: 'DataAnalysis' },
  { label: '监控', value: 'Monitor' }
]

// 菜单选项（用于父级菜单选择）
const menuOptions = computed(() => {
  const options = [{ id: 0, title: '顶级菜单', children: [] }]
  return options.concat(buildMenuOptions(menuList.value))
})

const buildMenuOptions = (menus) => {
  return menus.map(menu => ({
    id: menu.id,
    title: menu.title,
    children: menu.children ? buildMenuOptions(menu.children) : []
  }))
}

onMounted(() => {
  fetchMenus()
})

const fetchMenus = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/admin/menus')
    menuList.value = response.data || []
    menus.value = buildMenuTree(menuList.value)
  } catch (error) {
    ElMessage.error('获取菜单列表失败')
  } finally {
    loading.value = false
  }
}

const buildMenuTree = (menuList) => {
  const menuMap = new Map()
  const roots = []
  
  // 将所有菜单项放入map
  menuList.forEach(menu => {
    menuMap.set(menu.id, { ...menu, children: [] })
  })
  
  // 构建树形结构
  menuList.forEach(menu => {
    const menuItem = menuMap.get(menu.id)
    if (menu.parent_id === 0) {
      roots.push(menuItem)
    } else {
      const parent = menuMap.get(menu.parent_id)
      if (parent) {
        parent.children.push(menuItem)
      }
    }
  })
  
  return roots
}

const createMenu = async () => {
  if (!createFormRef.value) return
  
  const valid = await createFormRef.value.validate().catch(() => false)
  if (!valid) return

  createLoading.value = true
  try {
    await axios.post('/api/admin/menus', createForm.value)
    ElMessage.success('菜单创建成功')
    showCreateDialog.value = false
    
    // 重置表单
    createForm.value = {
      parent_id: 0,
      name: '',
      title: '',
      path: '',
      component: '',
      icon: '',
      sort: 0,
      type: 1
    }
    fetchMenus()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '创建菜单失败')
  } finally {
    createLoading.value = false
  }
}

const editMenu = (menu) => {
  editForm.value = {
    id: menu.id,
    parent_id: menu.parent_id,
    name: menu.name,
    title: menu.title,
    path: menu.path,
    component: menu.component,
    icon: menu.icon,
    sort: menu.sort,
    type: menu.type
  }
  showEditDialog.value = true
}

const updateMenu = async () => {
  if (!editFormRef.value) return
  
  const valid = await editFormRef.value.validate().catch(() => false)
  if (!valid) return

  editLoading.value = true
  try {
    await axios.put(`/api/admin/menus/${editForm.value.id}`, {
      parent_id: editForm.value.parent_id,
      name: editForm.value.name,
      title: editForm.value.title,
      path: editForm.value.path,
      component: editForm.value.component,
      icon: editForm.value.icon,
      sort: editForm.value.sort,
      type: editForm.value.type
    })
    ElMessage.success('菜单更新成功')
    showEditDialog.value = false
    fetchMenus()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '更新菜单失败')
  } finally {
    editLoading.value = false
  }
}

const toggleMenuStatus = async (menu) => {
  const newStatus = menu.status === 1 ? 0 : 1
  const action = newStatus === 1 ? '启用' : '禁用'
  
  try {
    await ElMessageBox.confirm(
      `确定要${action}菜单 ${menu.title} 吗？`,
      '确认操作',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    await axios.put(`/api/admin/menus/${menu.id}/status`, {
      status: newStatus
    })
    ElMessage.success(`${action}成功`)
    fetchMenus()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || `${action}失败`)
    }
  }
}

const deleteMenu = async (menu) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除菜单 ${menu.title} 吗？此操作不可逆！`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    await axios.delete(`/api/admin/menus/${menu.id}`)
    ElMessage.success('删除成功')
    fetchMenus()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '删除失败')
    }
  }
}
</script>

<style scoped>
.menu-management {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.icon-option {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style> 