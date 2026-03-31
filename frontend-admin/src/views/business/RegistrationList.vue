<template>
  <div class="registration-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>报名管理</span>
        </div>
      </template>

      <div class="filter-bar">
        <el-select v-model="filter.activity_id" placeholder="选择活动" clearable @change="loadRegistrations">
          <el-option label="全部活动" :value="0" />
          <el-option
            v-for="act in activities"
            :key="act.id"
            :label="act.title"
            :value="act.id"
          />
        </el-select>
        <el-select v-model="filter.status" placeholder="选择状态" @change="loadRegistrations">
          <el-option label="全部" :value="-1" />
          <el-option label="待支付" :value="0" />
          <el-option label="已支付" :value="1" />
          <el-option label="已取消" :value="2" />
          <el-option label="已退款" :value="3" />
        </el-select>
      </div>

      <el-table :data="registrations" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="activity_title" label="活动名称" show-overflow-tooltip />
        <el-table-column prop="name" label="报名人" width="120" />
        <el-table-column prop="phone" label="手机号" width="130" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="regStatusType(scope.row.status)">
              {{ regStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="报名时间" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="scope">
            <el-button size="small" @click="viewDetail(scope.row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.page_size"
        :total="pagination.total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadRegistrations"
        @current-change="loadRegistrations"
      />
    </el-card>

    <!-- 详情对话框 -->
    <el-dialog v-model="showDetail" title="报名详情" width="500px">
      <el-descriptions v-if="selectedReg" :column="1" border>
        <el-descriptions-item label="ID">{{ selectedReg.id }}</el-descriptions-item>
        <el-descriptions-item label="活动">{{ selectedReg.activity_title }}</el-descriptions-item>
        <el-descriptions-item label="报名人">{{ selectedReg.name }}</el-descriptions-item>
        <el-descriptions-item label="手机号">{{ selectedReg.phone }}</el-descriptions-item>
        <el-descriptions-item label="身份证">{{ selectedReg.id_card || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="regStatusType(selectedReg.status)">
            {{ regStatusText(selectedReg.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="报名时间">{{ formatDate(selectedReg.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="附加信息" v-if="selectedReg.extra_info">
          <pre class="extra-info">{{ JSON.stringify(selectedReg.extra_info, null, 2) }}</pre>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { registrationApi, activityApi } from '../../api'

const registrations = ref([])
const activities = ref([])
const loading = ref(false)
const showDetail = ref(false)
const selectedReg = ref(null)

const filter = ref({ activity_id: 0, status: -1 })
const pagination = ref({ page: 1, page_size: 20, total: 0 })

const regStatusMap = {
  0: { text: '待支付', type: 'warning' },
  1: { text: '已支付', type: 'success' },
  2: { text: '已取消', type: 'info' },
  3: { text: '已退款', type: 'danger' },
}
const regStatusText = (s) => regStatusMap[s]?.text || '未知'
const regStatusType = (s) => regStatusMap[s]?.type || 'info'

const loadActivities = async () => {
  try {
    const data = await activityApi.list({ page: 1, page_size: 100 })
    activities.value = data.list || []
  } catch (error) {
    console.error('加载活动列表失败:', error)
  }
}

const loadRegistrations = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.value.page,
      page_size: pagination.value.page_size,
    }
    if (filter.value.activity_id > 0) params.activity_id = filter.value.activity_id
    if (filter.value.status >= 0) params.status = filter.value.status

    const data = await registrationApi.list(params)
    registrations.value = data.list || []
    pagination.value.total = data.total
  } catch (error) {
    ElMessage.error('加载报名记录失败')
  } finally {
    loading.value = false
  }
}

const viewDetail = async (row) => {
  try {
    selectedReg.value = await registrationApi.get(row.id)
    showDetail.value = true
  } catch (error) {
    ElMessage.error('加载详情失败')
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

onMounted(() => {
  loadActivities()
  loadRegistrations()
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
.extra-info {
  font-size: 12px;
  background: #f5f7fa;
  padding: 8px;
  border-radius: 4px;
  max-height: 200px;
  overflow: auto;
}
</style>
