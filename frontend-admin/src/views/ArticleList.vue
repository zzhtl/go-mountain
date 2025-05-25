<template>
  <div class="article-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>文章管理</span>
          <el-button type="primary" @click="createArticle">新增文章</el-button>
        </div>
      </template>
      
      <!-- 筛选条件 -->
      <div class="filter-bar">
        <el-select v-model="filter.column_id" placeholder="选择栏目" clearable @change="loadArticles">
          <el-option label="全部栏目" :value="0" />
          <el-option
            v-for="column in columns"
            :key="column.id"
            :label="column.name"
            :value="column.id"
          />
        </el-select>
        
        <el-select v-model="filter.status" placeholder="选择状态" @change="loadArticles">
          <el-option label="全部" :value="-1" />
          <el-option label="草稿" :value="0" />
          <el-option label="已发布" :value="1" />
        </el-select>
      </div>
      
      <!-- 文章列表 -->
      <el-table :data="articles" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="title" label="标题" show-overflow-tooltip />
        <el-table-column prop="column_name" label="栏目" width="120" />
        <el-table-column prop="author" label="作者" width="100" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : 'info'">
              {{ scope.row.status === 1 ? '已发布' : '草稿' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="view_count" label="浏览量" width="100" />
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <el-button size="small" @click="editArticle(scope.row.id)">编辑</el-button>
            <el-button 
              size="small" 
              :type="scope.row.status === 1 ? 'warning' : 'success'"
              @click="toggleStatus(scope.row)"
            >
              {{ scope.row.status === 1 ? '下架' : '发布' }}
            </el-button>
            <el-button size="small" type="danger" @click="deleteArticle(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.page_size"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadArticles"
        @current-change="loadArticles"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import axios from 'axios'

const router = useRouter()
const articles = ref([])
const columns = ref([])
const loading = ref(false)

const filter = ref({
  column_id: 0,
  status: -1
})

const pagination = ref({
  page: 1,
  page_size: 20,
  total: 0
})

// 加载栏目列表
const loadColumns = async () => {
  try {
    const res = await axios.get('/api/admin/columns/')
    columns.value = res.data
  } catch (error) {
    console.error('加载栏目失败', error)
  }
}

// 加载文章列表
const loadArticles = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.value.page,
      page_size: pagination.value.page_size
    }
    
    if (filter.value.column_id > 0) {
      params.column_id = filter.value.column_id
    }
    
    if (filter.value.status >= 0) {
      params.status = filter.value.status
    }
    
    const res = await axios.get('/api/admin/articles/', { params })
    articles.value = res.data.list || []
    pagination.value.total = res.data.total
  } catch (error) {
    ElMessage.error('加载文章失败')
  } finally {
    loading.value = false
  }
}

// 创建文章
const createArticle = () => {
  router.push('/admin/articles/create')
}

// 编辑文章
const editArticle = (id) => {
  router.push(`/admin/articles/edit/${id}`)
}

// 切换文章状态
const toggleStatus = async (row) => {
  const newStatus = row.status === 1 ? 0 : 1
  const action = newStatus === 1 ? '发布' : '下架'
  
  try {
    await ElMessageBox.confirm(`确定要${action}该文章吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await axios.put(`/api/admin/articles/${row.id}/status`, { status: newStatus })
    ElMessage.success(`${action}成功`)
    loadArticles()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || `${action}失败`)
    }
  }
}

// 删除文章
const deleteArticle = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该文章吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await axios.delete(`/api/admin/articles/${row.id}`)
    ElMessage.success('删除成功')
    loadArticles()
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

onMounted(() => {
  loadColumns()
  loadArticles()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-bar {
  margin-bottom: 20px;
}

.filter-bar .el-select {
  margin-right: 10px;
}

.el-pagination {
  margin-top: 20px;
  text-align: right;
}
</style> 