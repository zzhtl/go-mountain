<template>
  <div class="column-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>栏目管理</span>
          <el-button type="primary" @click="showDialog = true">新增栏目</el-button>
        </div>
      </template>
      
      <el-table :data="columns" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="栏目名称" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="sort_order" label="排序" width="80" />
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180">
          <template #default="scope">
            <el-button size="small" @click="editColumn(scope.row)">编辑</el-button>
            <el-button size="small" type="danger" @click="deleteColumn(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="showDialog" :title="editingColumn ? '编辑栏目' : '新增栏目'" width="500px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="栏目名称" required>
          <el-input v-model="form.name" placeholder="请输入栏目名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" rows="3" placeholder="请输入栏目描述" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="form.sort_order" :min="0" :max="999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDialog = false">取消</el-button>
        <el-button type="primary" @click="saveColumn">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import axios from 'axios'

const columns = ref([])
const showDialog = ref(false)
const editingColumn = ref(null)
const form = ref({
  name: '',
  description: '',
  sort_order: 0
})

// 获取栏目列表
const loadColumns = async () => {
  try {
    const res = await axios.get('/api/admin/columns/')
    columns.value = res.data
  } catch (error) {
    ElMessage.error('加载栏目失败')
  }
}

// 新增/编辑栏目
const saveColumn = async () => {
  if (!form.value.name) {
    ElMessage.warning('请输入栏目名称')
    return
  }
  
  try {
    if (editingColumn.value) {
      await axios.put(`/api/admin/columns/${editingColumn.value.id}`, form.value)
      ElMessage.success('更新成功')
    } else {
      await axios.post('/api/admin/columns/', form.value)
      ElMessage.success('创建成功')
    }
    showDialog.value = false
    loadColumns()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '操作失败')
  }
}

// 编辑栏目
const editColumn = (row) => {
  editingColumn.value = row
  form.value = {
    name: row.name,
    description: row.description,
    sort_order: row.sort_order
  }
  showDialog.value = true
}

// 删除栏目
const deleteColumn = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该栏目吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await axios.delete(`/api/admin/columns/${row.id}`)
    ElMessage.success('删除成功')
    loadColumns()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '删除失败')
    }
  }
}

// 格式化日期
const formatDate = (dateStr) => {
  return new Date(dateStr).toLocaleString('zh-CN')
}

// 监听对话框关闭
const handleDialogClose = () => {
  editingColumn.value = null
  form.value = {
    name: '',
    description: '',
    sort_order: 0
  }
}

onMounted(() => {
  loadColumns()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style> 