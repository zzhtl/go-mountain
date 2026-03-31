<template>
  <div class="activity-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>活动管理</span>
          <el-button type="primary" @click="createActivity">新增活动</el-button>
        </div>
      </template>

      <div class="filter-bar">
        <el-select v-model="filter.status" placeholder="选择状态" @change="loadActivities">
          <el-option label="全部" :value="-1" />
          <el-option label="草稿" :value="0" />
          <el-option label="报名中" :value="1" />
          <el-option label="报名截止" :value="2" />
          <el-option label="进行中" :value="3" />
          <el-option label="已结束" :value="4" />
        </el-select>
      </div>

      <el-table :data="activities" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="title" label="活动名称" show-overflow-tooltip />
        <el-table-column prop="location" label="地点" width="150" show-overflow-tooltip />
        <el-table-column label="活动时间" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.start_time) }}
          </template>
        </el-table-column>
        <el-table-column prop="price" label="费用" width="100">
          <template #default="scope">
            {{ scope.row.price > 0 ? `¥${scope.row.price}` : '免费' }}
          </template>
        </el-table-column>
        <el-table-column label="报名人数" width="120">
          <template #default="scope">
            {{ scope.row.reg_count }}{{ scope.row.max_participants > 0 ? `/${scope.row.max_participants}` : '' }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="statusType(scope.row.status)">
              {{ statusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="scope">
            <el-button size="small" @click="editActivity(scope.row.id)">编辑</el-button>
            <el-button
              size="small"
              :type="scope.row.status === 1 ? 'warning' : 'success'"
              @click="toggleStatus(scope.row)"
              v-if="scope.row.status <= 1"
            >
              {{ scope.row.status === 1 ? '停止报名' : '开放报名' }}
            </el-button>
            <el-button size="small" type="danger" @click="deleteActivity(scope.row)" v-if="scope.row.status === 0">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.page_size"
        :total="pagination.total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadActivities"
        @current-change="loadActivities"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { activityApi } from '../../api'

const router = useRouter()
const activities = ref([])
const loading = ref(false)

const filter = ref({ status: -1 })
const pagination = ref({ page: 1, page_size: 20, total: 0 })

const statusMap = {
  0: { text: '草稿', type: 'info' },
  1: { text: '报名中', type: 'success' },
  2: { text: '报名截止', type: 'warning' },
  3: { text: '进行中', type: 'primary' },
  4: { text: '已结束', type: 'danger' },
}

const statusText = (s) => statusMap[s]?.text || '未知'
const statusType = (s) => statusMap[s]?.type || 'info'

const loadActivities = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.value.page,
      page_size: pagination.value.page_size,
    }
    if (filter.value.status >= 0) {
      params.status = filter.value.status
    }
    const data = await activityApi.list(params)
    activities.value = data.list || []
    pagination.value.total = data.total
  } catch (error) {
    ElMessage.error('加载活动失败')
  } finally {
    loading.value = false
  }
}

const createActivity = () => router.push('/admin/activities/create')
const editActivity = (id) => router.push(`/admin/activities/edit/${id}`)

const toggleStatus = async (row) => {
  const newStatus = row.status === 1 ? 2 : 1
  const action = newStatus === 1 ? '开放报名' : '停止报名'
  try {
    await ElMessageBox.confirm(`确定要${action}吗？`, '提示', { type: 'warning' })
    await activityApi.updateStatus(row.id, { status: newStatus })
    ElMessage.success(`${action}成功`)
    loadActivities()
  } catch (error) {
    if (error !== 'cancel') ElMessage.error(`${action}失败`)
  }
}

const deleteActivity = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该活动吗？', '提示', { type: 'warning' })
    await activityApi.delete(row.id)
    ElMessage.success('删除成功')
    loadActivities()
  } catch (error) {
    if (error !== 'cancel') ElMessage.error('删除失败')
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

onMounted(() => loadActivities())
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
.el-pagination {
  margin-top: 20px;
  text-align: right;
}
</style>
