<template>
  <el-card v-loading="initialLoading">
    <template #header>
      <div style="display: flex; justify-content: space-between; align-items: center">
        <span>{{ config.display_name || '数据管理' }}</span>
        <el-button type="primary" @click="openForm()">新增</el-button>
      </div>
    </template>

    <!-- 搜索栏 -->
    <div v-if="searchColumns.length" style="margin-bottom: 16px; display: flex; gap: 12px; align-items: center">
      <el-input
        v-model="keyword"
        placeholder="搜索..."
        clearable
        style="width: 300px"
        @keyup.enter="loadData"
        @clear="loadData"
      />
      <el-button type="primary" @click="loadData">搜索</el-button>
    </div>

    <!-- 数据表格 -->
    <el-table :data="list" stripe v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" sortable />
      <template v-for="col in listColumns" :key="col.field">
        <el-table-column :prop="col.field" :label="col.label" :sortable="col.sortable" :min-width="colWidth(col)">
          <template #default="{ row }">
            <!-- 图片类型 -->
            <el-image
              v-if="col.form_type === 'image' && row[col.field]"
              :src="row[col.field]"
              style="width: 60px; height: 60px; object-fit: cover"
              :preview-src-list="[row[col.field]]"
              fit="cover"
            />
            <!-- 开关类型 -->
            <el-tag v-else-if="col.form_type === 'switch'" :type="row[col.field] ? 'success' : 'info'">
              {{ row[col.field] ? '是' : '否' }}
            </el-tag>
            <!-- 下拉选项显示标签 -->
            <span v-else-if="col.form_type === 'select' && col.options?.length">
              {{ getOptionLabel(col, row[col.field]) }}
            </span>
            <!-- 长文本截断 -->
            <span v-else-if="col.form_type === 'textarea' || col.form_type === 'richtext'">
              {{ truncate(row[col.field], 50) }}
            </span>
            <!-- 日期格式化 -->
            <span v-else-if="col.form_type === 'date' || col.form_type === 'datetime'">
              {{ formatDate(row[col.field]) }}
            </span>
            <!-- 默认 -->
            <span v-else>{{ row[col.field] }}</span>
          </template>
        </el-table-column>
      </template>
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="160">
        <template #default="{ row }">
          <el-button size="small" @click="openForm(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div style="margin-top: 16px; display: flex; justify-content: flex-end">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @current-change="loadData"
        @size-change="loadData"
      />
    </div>

    <!-- 新增/编辑弹窗 -->
    <el-dialog
      v-model="formVisible"
      :title="editingId ? '编辑' : '新增'"
      width="640px"
      @closed="resetForm"
    >
      <el-form :model="formData" label-width="120px">
        <template v-for="col in columns" :key="col.field">
          <el-form-item :label="col.label" :required="col.required">
            <!-- 文本框 -->
            <el-input v-if="col.form_type === 'input'" v-model="formData[col.field]" />
            <!-- 文本域 -->
            <el-input v-else-if="col.form_type === 'textarea'" v-model="formData[col.field]" type="textarea" :rows="3" />
            <!-- 数字 -->
            <el-input-number v-else-if="col.form_type === 'number'" v-model="formData[col.field]" style="width: 100%" />
            <!-- 下拉框 -->
            <el-select v-else-if="col.form_type === 'select'" v-model="formData[col.field]" clearable style="width: 100%">
              <el-option
                v-for="opt in (col.options || [])"
                :key="opt.value"
                :label="opt.label"
                :value="opt.value"
              />
            </el-select>
            <!-- 日期 -->
            <el-date-picker
              v-else-if="col.form_type === 'date'"
              v-model="formData[col.field]"
              type="date"
              style="width: 100%"
            />
            <!-- 日期时间 -->
            <el-date-picker
              v-else-if="col.form_type === 'datetime'"
              v-model="formData[col.field]"
              type="datetime"
              style="width: 100%"
            />
            <!-- 图片上传 -->
            <div v-else-if="col.form_type === 'image'">
              <el-input v-model="formData[col.field]" placeholder="图片 URL">
                <template #append>
                  <el-upload
                    :action="uploadUrl"
                    :headers="uploadHeaders"
                    :show-file-list="false"
                    :on-success="(res) => { formData[col.field] = res.data?.url || '' }"
                  >
                    <el-button>上传</el-button>
                  </el-upload>
                </template>
              </el-input>
              <el-image
                v-if="formData[col.field]"
                :src="formData[col.field]"
                style="width: 100px; height: 100px; margin-top: 8px"
                fit="cover"
              />
            </div>
            <!-- 开关 -->
            <el-switch v-else-if="col.form_type === 'switch'" v-model="formData[col.field]" />
            <!-- 富文本（简化为 textarea，如需可替换为 RichEditor） -->
            <el-input v-else-if="col.form_type === 'richtext'" v-model="formData[col.field]" type="textarea" :rows="6" />
            <!-- 兜底 -->
            <el-input v-else v-model="formData[col.field]" />
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import request from '../../api/request'

const route = useRoute()

// 从路由 meta 或路径推断模块路由名
const routeName = computed(() => {
  const path = route.path.replace(/^\/admin\//, '')
  return path
})

// API 基础路径
const apiBase = computed(() => `/api/admin/${routeName.value}/`)

// 配置信息
const config = ref({ display_name: '', columns: [] })
const columns = ref([])
const listColumns = computed(() => columns.value.filter(c => c.list_visible))
const searchColumns = computed(() => columns.value.filter(c => c.searchable))

// 数据
const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const keyword = ref('')
const loading = ref(false)
const initialLoading = ref(true)

// 表单
const formVisible = ref(false)
const formData = ref({})
const editingId = ref(null)
const saving = ref(false)

// 上传相关
const uploadUrl = '/api/admin/upload/image'
const uploadHeaders = computed(() => ({
  Authorization: `Bearer ${localStorage.getItem('token')}`
}))

// 加载 codegen 配置
const loadConfig = async () => {
  try {
    // 通过模块路由名查找对应的 codegen 配置
    const data = await request.get('/api/admin/codegen/', { params: { page: 1, page_size: 100 } })
    const configs = data.list || []
    // 根据路由名匹配配置
    const moduleName = routeName.value.replace(/^gen-/, '').replace(/-/g, '_')
    const found = configs.find(c => c.module_name === moduleName)
    if (found) {
      config.value = found
      const parsed = typeof found.columns_config === 'string'
        ? JSON.parse(found.columns_config)
        : found.columns_config
      columns.value = parsed || []
    }
  } catch (e) {
    console.error('加载配置失败', e)
  }
}

// 加载数据
const loadData = async () => {
  loading.value = true
  try {
    const params = { page: page.value, page_size: pageSize.value }
    if (keyword.value) params.keyword = keyword.value
    const data = await request.get(apiBase.value, { params })
    list.value = data.list || []
    total.value = data.total || 0
  } catch (e) {
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

// 打开表单
const openForm = (row) => {
  if (row) {
    editingId.value = row.id
    formData.value = { ...row }
  } else {
    editingId.value = null
    formData.value = {}
    // 设置默认值
    columns.value.forEach(col => {
      if (col.form_type === 'switch') formData.value[col.field] = false
      else if (col.form_type === 'number') formData.value[col.field] = 0
      else formData.value[col.field] = ''
    })
  }
  formVisible.value = true
}

const resetForm = () => {
  formData.value = {}
  editingId.value = null
}

// 保存
const handleSave = async () => {
  saving.value = true
  try {
    if (editingId.value) {
      await request.put(`${apiBase.value}${editingId.value}`, formData.value)
      ElMessage.success('更新成功')
    } else {
      await request.post(apiBase.value, formData.value)
      ElMessage.success('创建成功')
    }
    formVisible.value = false
    loadData()
  } catch (e) {
    ElMessage.error('保存失败: ' + (e.message || '未知错误'))
  } finally {
    saving.value = false
  }
}

// 删除
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除吗？', '确认删除', { type: 'warning' })
    await request.delete(`${apiBase.value}${row.id}`)
    ElMessage.success('删除成功')
    loadData()
  } catch (e) {
    if (e !== 'cancel') ElMessage.error('删除失败')
  }
}

// 辅助
const getOptionLabel = (col, value) => {
  const opt = (col.options || []).find(o => String(o.value) === String(value))
  return opt ? opt.label : value
}

const truncate = (str, len) => {
  if (!str) return ''
  str = String(str).replace(/<[^>]+>/g, '') // 移除 HTML 标签
  return str.length > len ? str.slice(0, len) + '...' : str
}

const formatDate = (str) => {
  if (!str) return ''
  return new Date(str).toLocaleString('zh-CN')
}

const colWidth = (col) => {
  if (col.form_type === 'image') return 100
  if (col.form_type === 'textarea' || col.form_type === 'richtext') return 200
  return 150
}

onMounted(async () => {
  await loadConfig()
  await loadData()
  initialLoading.value = false
})
</script>
