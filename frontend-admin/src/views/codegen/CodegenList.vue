<template>
  <el-card>
    <template #header>
      <div style="display: flex; justify-content: space-between; align-items: center">
        <span>代码生成配置</span>
        <el-button type="primary" @click="$router.push('/admin/codegen/create')">新建配置</el-button>
      </div>
    </template>

    <el-table :data="list" stripe v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="table_name" label="数据库表" width="180" />
      <el-table-column prop="module_name" label="模块名" width="150" />
      <el-table-column prop="display_name" label="显示名称" width="150" />
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="row.generated ? 'success' : 'info'">
            {{ row.generated ? '已生成' : '未生成' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="生成时间" width="180">
        <template #default="{ row }">
          {{ row.generated_at ? formatDate(row.generated_at) : '-' }}
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
      </el-table-column>
      <el-table-column label="操作" fixed="right" width="280">
        <template #default="{ row }">
          <el-button size="small" @click="$router.push(`/admin/codegen/edit/${row.id}`)">编辑</el-button>
          <el-button size="small" type="warning" @click="handlePreview(row)">预览</el-button>
          <el-button size="small" type="success" @click="handleGenerate(row)">
            {{ row.generated ? '重新生成' : '生成代码' }}
          </el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div style="margin-top: 16px; display: flex; justify-content: flex-end">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @current-change="loadList"
        @size-change="loadList"
      />
    </div>

    <!-- 代码预览弹窗 -->
    <el-dialog v-model="previewVisible" title="代码预览" width="80%" top="5vh">
      <el-tabs v-model="previewTab">
        <el-tab-pane label="Model" name="model">
          <pre class="code-block">{{ previewCode.model_code }}</pre>
        </el-tab-pane>
        <el-tab-pane label="Service" name="service">
          <pre class="code-block">{{ previewCode.service_code }}</pre>
        </el-tab-pane>
        <el-tab-pane label="Handler" name="handler">
          <pre class="code-block">{{ previewCode.handler_code }}</pre>
        </el-tab-pane>
        <el-tab-pane label="Router 代码片段" name="router">
          <pre class="code-block">{{ previewCode.router_code }}</pre>
        </el-tab-pane>
        <el-tab-pane label="API 代码片段" name="api">
          <pre class="code-block">{{ previewCode.api_code }}</pre>
        </el-tab-pane>
      </el-tabs>
    </el-dialog>
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { codegenApi } from '../../api'

const list = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

const previewVisible = ref(false)
const previewTab = ref('model')
const previewCode = ref({})

const loadList = async () => {
  loading.value = true
  try {
    const data = await codegenApi.list({ page: page.value, page_size: pageSize.value })
    list.value = data.list || []
    total.value = data.total || 0
  } finally {
    loading.value = false
  }
}

const handlePreview = async (row) => {
  try {
    const data = await codegenApi.preview(row.id)
    previewCode.value = data
    previewTab.value = 'model'
    previewVisible.value = true
  } catch (e) {
    ElMessage.error('预览失败: ' + (e.message || '未知错误'))
  }
}

const handleGenerate = async (row) => {
  const action = row.generated ? '重新生成（将覆盖已有文件）' : '生成代码'
  try {
    await ElMessageBox.confirm(
      `确定要${action}吗？生成后需要重新编译后端并重启服务。`,
      '确认生成',
      { type: 'warning' }
    )
    const data = await codegenApi.generate(row.id)
    previewCode.value = data
    previewTab.value = 'router'
    previewVisible.value = true
    ElMessage.success('代码生成成功！请将 Router 和 API 代码片段手动添加到对应文件中，然后重新编译。')
    loadList()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('生成失败: ' + (e.message || '未知错误'))
    }
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该配置吗？', '确认删除', { type: 'warning' })
    await codegenApi.delete(row.id)
    ElMessage.success('删除成功')
    loadList()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const formatDate = (str) => {
  if (!str) return ''
  return new Date(str).toLocaleString('zh-CN')
}

onMounted(loadList)
</script>

<style scoped>
.code-block {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 16px;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 13px;
  line-height: 1.5;
  max-height: 60vh;
  white-space: pre;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
}
</style>
