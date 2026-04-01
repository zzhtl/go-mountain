<template>
  <el-card>
    <template #header>
      <div style="display: flex; justify-content: space-between; align-items: center">
        <span>{{ isEdit ? '编辑生成配置' : '新建生成配置' }}</span>
        <el-button @click="$router.back()">返回</el-button>
      </div>
    </template>

    <!-- 步骤条 -->
    <el-steps :active="step" finish-status="success" style="margin-bottom: 24px">
      <el-step title="选择数据表" />
      <el-step title="配置字段" />
      <el-step title="确认保存" />
    </el-steps>

    <!-- 第一步：选择表 + 基本信息 -->
    <div v-show="step === 0">
      <el-form label-width="120px" style="max-width: 600px">
        <el-form-item label="数据库表" required>
          <el-select
            v-model="form.table_name"
            filterable
            placeholder="选择数据库表"
            @change="onTableChange"
            style="width: 100%"
          >
            <el-option
              v-for="t in tables"
              :key="t.table_name"
              :label="t.table_name + (t.comment ? ' - ' + t.comment : '')"
              :value="t.table_name"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="模块名" required>
          <el-input v-model="form.module_name" placeholder="如 product（snake_case，用于文件名和路由）" />
        </el-form-item>
        <el-form-item label="显示名称" required>
          <el-input v-model="form.display_name" placeholder="如 商品管理（显示在菜单和页面标题）" />
        </el-form-item>
        <el-form-item label="父级菜单">
          <el-select v-model="form.parent_menu_id" clearable placeholder="不选则自动创建目录" style="width: 100%">
            <el-option
              v-for="m in parentMenus"
              :key="m.id"
              :label="m.title"
              :value="m.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <div style="text-align: right; margin-top: 16px">
        <el-button type="primary" @click="nextStep" :disabled="!form.table_name || !form.module_name || !form.display_name">
          下一步
        </el-button>
      </div>
    </div>

    <!-- 第二步：配置字段 -->
    <div v-show="step === 1">
      <el-alert type="info" :closable="false" style="margin-bottom: 16px">
        配置每个字段在前端页面中的展示方式。id、created_at、updated_at、deleted_at 已自动排除（由 BaseModel 管理）。
      </el-alert>

      <el-table :data="form.columns" stripe border>
        <el-table-column prop="field" label="字段名" width="140" />
        <el-table-column label="显示名称" width="150">
          <template #default="{ row }">
            <el-input v-model="row.label" size="small" />
          </template>
        </el-table-column>
        <el-table-column prop="type" label="Go 类型" width="130" />
        <el-table-column label="表单类型" width="140">
          <template #default="{ row }">
            <el-select v-model="row.form_type" size="small">
              <el-option label="文本框" value="input" />
              <el-option label="文本域" value="textarea" />
              <el-option label="数字" value="number" />
              <el-option label="下拉框" value="select" />
              <el-option label="日期" value="date" />
              <el-option label="日期时间" value="datetime" />
              <el-option label="图片上传" value="image" />
              <el-option label="富文本" value="richtext" />
              <el-option label="开关" value="switch" />
            </el-select>
          </template>
        </el-table-column>
        <el-table-column label="列表显示" width="90" align="center">
          <template #default="{ row }">
            <el-checkbox v-model="row.list_visible" />
          </template>
        </el-table-column>
        <el-table-column label="可搜索" width="80" align="center">
          <template #default="{ row }">
            <el-checkbox v-model="row.searchable" />
          </template>
        </el-table-column>
        <el-table-column label="可排序" width="80" align="center">
          <template #default="{ row }">
            <el-checkbox v-model="row.sortable" />
          </template>
        </el-table-column>
        <el-table-column label="必填" width="70" align="center">
          <template #default="{ row }">
            <el-checkbox v-model="row.required" />
          </template>
        </el-table-column>
        <el-table-column label="下拉选项" min-width="200">
          <template #default="{ row }">
            <div v-if="row.form_type === 'select'">
              <el-tag
                v-for="(opt, idx) in (row.options || [])"
                :key="idx"
                closable
                size="small"
                style="margin-right: 4px; margin-bottom: 4px"
                @close="row.options.splice(idx, 1)"
              >
                {{ opt.label }}={{ opt.value }}
              </el-tag>
              <el-popover trigger="click" width="240">
                <template #reference>
                  <el-button size="small" type="primary" link>+ 添加选项</el-button>
                </template>
                <el-form size="small">
                  <el-form-item label="标签">
                    <el-input v-model="newOption.label" />
                  </el-form-item>
                  <el-form-item label="值">
                    <el-input v-model="newOption.value" />
                  </el-form-item>
                  <el-button size="small" type="primary" @click="addOption(row)">确定</el-button>
                </el-form>
              </el-popover>
            </div>
            <span v-else style="color: #999">-</span>
          </template>
        </el-table-column>
      </el-table>

      <div style="text-align: right; margin-top: 16px">
        <el-button @click="step = 0">上一步</el-button>
        <el-button type="primary" @click="step = 2">下一步</el-button>
      </div>
    </div>

    <!-- 第三步：确认保存 -->
    <div v-show="step === 2">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="数据库表">{{ form.table_name }}</el-descriptions-item>
        <el-descriptions-item label="模块名">{{ form.module_name }}</el-descriptions-item>
        <el-descriptions-item label="显示名称">{{ form.display_name }}</el-descriptions-item>
        <el-descriptions-item label="字段数量">{{ form.columns.length }}</el-descriptions-item>
        <el-descriptions-item label="列表显示字段">
          {{ form.columns.filter(c => c.list_visible).map(c => c.label).join(', ') || '无' }}
        </el-descriptions-item>
        <el-descriptions-item label="可搜索字段">
          {{ form.columns.filter(c => c.searchable).map(c => c.label).join(', ') || '无' }}
        </el-descriptions-item>
      </el-descriptions>

      <div style="text-align: right; margin-top: 24px">
        <el-button @click="step = 1">上一步</el-button>
        <el-button type="primary" :loading="saving" @click="handleSave">
          {{ isEdit ? '保存配置' : '创建配置' }}
        </el-button>
      </div>
    </div>
  </el-card>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { codegenApi, menuApi } from '../../api'

const router = useRouter()
const route = useRoute()
const isEdit = computed(() => !!route.params.id)

const step = ref(0)
const tables = ref([])
const parentMenus = ref([])
const saving = ref(false)

const form = ref({
  table_name: '',
  module_name: '',
  display_name: '',
  parent_menu_id: 0,
  columns: []
})

const newOption = ref({ label: '', value: '' })

const addOption = (row) => {
  if (!row.options) row.options = []
  if (newOption.value.label && newOption.value.value !== '') {
    row.options.push({ ...newOption.value })
    newOption.value = { label: '', value: '' }
  }
}

// 选择表后自动读取列信息并推断模块名
const onTableChange = async (tableName) => {
  if (!tableName) return

  // 自动推断模块名和显示名
  if (!form.value.module_name) {
    // 去除复数 s 作为模块名
    let name = tableName
    if (name.endsWith('ies')) {
      name = name.slice(0, -3) + 'y'
    } else if (name.endsWith('ses') || name.endsWith('xes') || name.endsWith('zes')) {
      name = name.slice(0, -2)
    } else if (name.endsWith('s') && !name.endsWith('ss')) {
      name = name.slice(0, -1)
    }
    form.value.module_name = name
  }

  try {
    const columns = await codegenApi.getColumns(tableName)
    form.value.columns = (columns || []).map(col => ({
      field: col.field,
      label: col.comment || col.go_field,
      type: col.go_type,
      form_type: col.form_type,
      list_visible: true,
      searchable: col.form_type === 'input',
      sortable: false,
      required: !col.nullable,
      options: []
    }))
  } catch (e) {
    ElMessage.error('读取表结构失败')
  }
}

const nextStep = () => {
  if (form.value.columns.length === 0 && form.value.table_name) {
    onTableChange(form.value.table_name).then(() => { step.value = 1 })
  } else {
    step.value = 1
  }
}

const handleSave = async () => {
  saving.value = true
  try {
    const payload = {
      table_name: form.value.table_name,
      module_name: form.value.module_name,
      display_name: form.value.display_name,
      parent_menu_id: form.value.parent_menu_id || 0,
      columns: form.value.columns
    }

    if (isEdit.value) {
      await codegenApi.update(route.params.id, payload)
      ElMessage.success('配置已更新')
    } else {
      await codegenApi.create(payload)
      ElMessage.success('配置已创建')
    }
    router.push('/admin/codegen')
  } catch (e) {
    ElMessage.error('保存失败: ' + (e.message || '未知错误'))
  } finally {
    saving.value = false
  }
}

// 加载初始数据
onMounted(async () => {
  // 加载数据库表列表
  try {
    tables.value = await codegenApi.getTables() || []
  } catch (e) {
    ElMessage.error('获取数据库表列表失败')
  }

  // 加载目录菜单（type=1）作为可选父菜单
  try {
    const allMenus = await menuApi.list()
    parentMenus.value = (allMenus || []).filter(m => m.type === 1)
  } catch (e) {
    // 忽略
  }

  // 编辑模式：加载已有配置
  if (isEdit.value) {
    try {
      const data = await codegenApi.get(route.params.id)
      form.value.table_name = data.table_name
      form.value.module_name = data.module_name
      form.value.display_name = data.display_name
      form.value.parent_menu_id = data.parent_menu_id || 0
      form.value.columns = JSON.parse(
        typeof data.columns_config === 'string' ? data.columns_config : JSON.stringify(data.columns_config)
      )
    } catch (e) {
      ElMessage.error('加载配置失败')
    }
  }
})
</script>
