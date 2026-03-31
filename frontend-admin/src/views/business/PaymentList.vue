<template>
  <div class="payment-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>支付管理</span>
        </div>
      </template>

      <div class="filter-bar">
        <el-select v-model="filter.status" placeholder="支付状态" @change="loadPayments">
          <el-option label="全部" :value="-1" />
          <el-option label="待支付" :value="0" />
          <el-option label="已支付" :value="1" />
          <el-option label="已退款" :value="2" />
          <el-option label="支付失败" :value="3" />
        </el-select>
        <el-select v-model="filter.biz_type" placeholder="业务类型" clearable @change="loadPayments">
          <el-option label="全部" value="" />
          <el-option label="报名" value="registration" />
          <el-option label="捐赠" value="donation" />
        </el-select>
      </div>

      <el-table :data="payments" stripe v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="order_no" label="订单号" width="200" />
        <el-table-column prop="amount" label="金额" width="100">
          <template #default="scope">
            ¥{{ scope.row.amount }}
          </template>
        </el-table-column>
        <el-table-column prop="pay_type" label="支付方式" width="120">
          <template #default="scope">
            {{ scope.row.pay_type === 'wechat_jsapi' ? '微信支付' : scope.row.pay_type }}
          </template>
        </el-table-column>
        <el-table-column prop="biz_type" label="业务类型" width="100">
          <template #default="scope">
            {{ scope.row.biz_type === 'registration' ? '报名' : scope.row.biz_type }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="payStatusType(scope.row.status)">
              {{ payStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="scope">
            {{ formatDate(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column prop="paid_at" label="支付时间" width="180">
          <template #default="scope">
            {{ scope.row.paid_at ? formatDate(scope.row.paid_at) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="scope">
            <el-button size="small" @click="viewDetail(scope.row)">详情</el-button>
            <el-button
              size="small"
              type="danger"
              @click="refundPayment(scope.row)"
              v-if="scope.row.status === 1"
            >
              退款
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.page_size"
        :total="pagination.total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadPayments"
        @current-change="loadPayments"
      />
    </el-card>

    <!-- 详情对话框 -->
    <el-dialog v-model="showDetail" title="支付详情" width="500px">
      <el-descriptions v-if="selectedPayment" :column="1" border>
        <el-descriptions-item label="ID">{{ selectedPayment.id }}</el-descriptions-item>
        <el-descriptions-item label="订单号">{{ selectedPayment.order_no }}</el-descriptions-item>
        <el-descriptions-item label="交易号">{{ selectedPayment.transaction_id || '-' }}</el-descriptions-item>
        <el-descriptions-item label="金额">¥{{ selectedPayment.amount }}</el-descriptions-item>
        <el-descriptions-item label="支付方式">{{ selectedPayment.pay_type === 'wechat_jsapi' ? '微信支付' : selectedPayment.pay_type }}</el-descriptions-item>
        <el-descriptions-item label="业务类型">{{ selectedPayment.biz_type }}</el-descriptions-item>
        <el-descriptions-item label="业务ID">{{ selectedPayment.biz_id }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="payStatusType(selectedPayment.status)">
            {{ payStatusText(selectedPayment.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDate(selectedPayment.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="支付时间">{{ selectedPayment.paid_at ? formatDate(selectedPayment.paid_at) : '-' }}</el-descriptions-item>
        <el-descriptions-item label="退款时间">{{ selectedPayment.refund_at ? formatDate(selectedPayment.refund_at) : '-' }}</el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { paymentApi } from '../../api'

const payments = ref([])
const loading = ref(false)
const showDetail = ref(false)
const selectedPayment = ref(null)

const filter = ref({ status: -1, biz_type: '' })
const pagination = ref({ page: 1, page_size: 20, total: 0 })

const payStatusMap = {
  0: { text: '待支付', type: 'warning' },
  1: { text: '已支付', type: 'success' },
  2: { text: '已退款', type: 'danger' },
  3: { text: '支付失败', type: 'info' },
}
const payStatusText = (s) => payStatusMap[s]?.text || '未知'
const payStatusType = (s) => payStatusMap[s]?.type || 'info'

const loadPayments = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.value.page,
      page_size: pagination.value.page_size,
    }
    if (filter.value.status >= 0) params.status = filter.value.status
    if (filter.value.biz_type) params.biz_type = filter.value.biz_type

    const data = await paymentApi.list(params)
    payments.value = data.list || []
    pagination.value.total = data.total
  } catch (error) {
    ElMessage.error('加载支付记录失败')
  } finally {
    loading.value = false
  }
}

const viewDetail = async (row) => {
  try {
    selectedPayment.value = await paymentApi.get(row.id)
    showDetail.value = true
  } catch (error) {
    ElMessage.error('加载详情失败')
  }
}

const refundPayment = async (row) => {
  try {
    await ElMessageBox.confirm(`确定要退款订单 ${row.order_no} 吗？金额 ¥${row.amount}`, '确认退款', { type: 'warning' })
    await paymentApi.refund(row.id)
    ElMessage.success('退款成功')
    loadPayments()
  } catch (error) {
    if (error !== 'cancel') ElMessage.error('退款失败')
  }
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

onMounted(() => loadPayments())
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
