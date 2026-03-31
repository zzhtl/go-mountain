<template>
  <div class="system-config">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>系统配置</span>
          <div>
            <el-button type="primary" @click="saveAll" :loading="saving">保存所有修改</el-button>
            <el-button @click="showAddDialog = true">新增配置</el-button>
          </div>
        </div>
      </template>

      <!-- 分组 Tab -->
      <el-tabs v-model="activeGroup" @tab-click="onTabClick">
        <el-tab-pane label="全部" name="all" />
        <el-tab-pane
          v-for="group in groups"
          :key="group"
          :label="group"
          :name="group"
        />
      </el-tabs>

      <!-- 配置列表 -->
      <el-table :data="filteredConfigs" stripe v-loading="loading">
        <el-table-column prop="key" label="配置项" width="250">
          <template #default="scope">
            <div>
              <span class="config-key">{{ scope.row.key }}</span>
              <div class="config-remark" v-if="scope.row.remark">{{ scope.row.remark }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="group_name" label="分组" width="120" />
        <el-table-column prop="value" label="配置值">
          <template #default="scope">
            <el-input
              v-if="scope.row.type === 'string' || !scope.row.type"
              v-model="scope.row.value"
              :type="isSecretKey(scope.row.key) ? 'password' : 'text'"
              :show-password="isSecretKey(scope.row.key)"
              @change="markDirty(scope.row)"
            />
            <el-input
              v-else-if="scope.row.type === 'json'"
              v-model="scope.row.value"
              type="textarea"
              :rows="3"
              @change="markDirty(scope.row)"
            />
            <el-input-number
              v-else-if="scope.row.type === 'number'"
              v-model.number="scope.row.value"
              @change="markDirty(scope.row)"
            />
            <el-switch
              v-else-if="scope.row.type === 'boolean'"
              v-model="scope.row.value"
              active-value="true"
              inactive-value="false"
              @change="markDirty(scope.row)"
            />
          </template>
        </el-table-column>
        <el-table-column label="状态" width="80">
          <template #default="scope">
            <el-tag v-if="scope.row._dirty" type="warning" size="small">已修改</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="scope">
            <el-button size="small" type="danger" @click="deleteConfig(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新增配置对话框 -->
    <el-dialog v-model="showAddDialog" title="新增配置项" width="500px">
      <el-form :model="addForm" label-width="100px">
        <el-form-item label="配置键" required>
          <el-input v-model="addForm.key" placeholder="如 wechat.mch_id" />
        </el-form-item>
        <el-form-item label="配置值">
          <el-input v-model="addForm.value" placeholder="请输入配置值" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="addForm.type">
            <el-option label="字符串" value="string" />
            <el-option label="数字" value="number" />
            <el-option label="布尔" value="boolean" />
            <el-option label="JSON" value="json" />
          </el-select>
        </el-form-item>
        <el-form-item label="分组">
          <el-select v-model="addForm.group_name" filterable allow-create placeholder="选择或输入分组">
            <el-option v-for="g in groups" :key="g" :label="g" :value="g" />
          </el-select>
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="addForm.remark" placeholder="配置说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="addConfig">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { systemConfigApi } from '../../api'

const configs = ref([])
const groups = ref([])
const loading = ref(false)
const saving = ref(false)
const activeGroup = ref('all')
const showAddDialog = ref(false)

const addForm = ref({
  key: '',
  value: '',
  type: 'string',
  group_name: '',
  remark: ''
})

// 敏感配置键（显示为密码框）
const secretKeywords = ['secret', 'password', 'private_key', 'api_key', 'api_v3_key']
const isSecretKey = (key) => secretKeywords.some(kw => key.toLowerCase().includes(kw))

const filteredConfigs = computed(() => {
  if (activeGroup.value === 'all') return configs.value
  return configs.value.filter(c => c.group_name === activeGroup.value)
})

const loadConfigs = async () => {
  loading.value = true
  try {
    const [configList, groupList] = await Promise.all([
      systemConfigApi.list(),
      systemConfigApi.groups()
    ])
    configs.value = (configList || []).map(c => ({ ...c, _dirty: false }))
    groups.value = groupList || []
  } catch (error) {
    ElMessage.error('加载配置失败')
  } finally {
    loading.value = false
  }
}

const markDirty = (row) => {
  row._dirty = true
}

const onTabClick = () => {
  // tab 切换不需要重新加载
}

const saveAll = async () => {
  const dirtyConfigs = configs.value.filter(c => c._dirty)
  if (dirtyConfigs.length === 0) {
    ElMessage.info('没有需要保存的修改')
    return
  }

  saving.value = true
  try {
    await systemConfigApi.batchSave({
      configs: dirtyConfigs.map(c => ({
        key: c.key,
        value: String(c.value),
        type: c.type,
        group_name: c.group_name,
        remark: c.remark
      }))
    })
    dirtyConfigs.forEach(c => c._dirty = false)
    ElMessage.success(`已保存 ${dirtyConfigs.length} 项配置`)
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

const addConfig = async () => {
  if (!addForm.value.key) {
    ElMessage.warning('请输入配置键')
    return
  }
  try {
    await systemConfigApi.save(addForm.value)
    ElMessage.success('配置已添加')
    showAddDialog.value = false
    addForm.value = { key: '', value: '', type: 'string', group_name: '', remark: '' }
    loadConfigs()
  } catch (error) {
    // 错误已在拦截器处理
  }
}

const deleteConfig = async (row) => {
  try {
    await ElMessageBox.confirm(`确定删除配置项 "${row.key}" 吗？`, '提示', { type: 'warning' })
    await systemConfigApi.delete(row.key)
    ElMessage.success('删除成功')
    loadConfigs()
  } catch (error) {
    if (error !== 'cancel') ElMessage.error('删除失败')
  }
}

onMounted(() => loadConfigs())
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.config-key {
  font-family: monospace;
  font-weight: 600;
  color: #409eff;
}
.config-remark {
  font-size: 12px;
  color: #999;
  margin-top: 2px;
}
</style>
