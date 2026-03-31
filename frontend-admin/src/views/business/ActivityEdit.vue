<template>
  <div class="activity-edit">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ isEdit ? '编辑活动' : '新增活动' }}</span>
          <div>
            <el-button @click="$router.back()">返回</el-button>
            <el-button type="primary" @click="saveActivity(0)">保存草稿</el-button>
            <el-button type="success" @click="saveActivity(1)">保存并开放报名</el-button>
          </div>
        </div>
      </template>

      <el-form :model="form" :rules="rules" ref="formRef" label-width="120px" class="activity-form">
        <el-form-item label="活动名称" prop="title">
          <el-input v-model="form.title" placeholder="请输入活动名称" />
        </el-form-item>

        <el-form-item label="活动简介" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入活动简介" />
        </el-form-item>

        <el-form-item label="缩略图">
          <div class="thumbnail-upload">
            <el-upload
              :action="uploadUrl"
              :headers="uploadHeaders"
              :show-file-list="false"
              :on-success="handleThumbnailSuccess"
              accept="image/*"
            >
              <img v-if="form.thumbnail" :src="form.thumbnail" class="thumbnail" />
              <el-icon v-else class="upload-icon"><Plus /></el-icon>
            </el-upload>
            <el-button v-if="form.thumbnail" type="danger" size="small" @click="form.thumbnail = ''">删除</el-button>
          </div>
        </el-form-item>

        <el-form-item label="活动地点">
          <el-input v-model="form.location" placeholder="请输入活动地点" />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="活动开始时间" prop="start_time">
              <el-date-picker
                v-model="form.start_time"
                type="datetime"
                placeholder="选择开始时间"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="活动结束时间" prop="end_time">
              <el-date-picker
                v-model="form.end_time"
                type="datetime"
                placeholder="选择结束时间"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="报名开始时间">
              <el-date-picker
                v-model="form.reg_start_time"
                type="datetime"
                placeholder="选择报名开始时间"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="报名截止时间">
              <el-date-picker
                v-model="form.reg_end_time"
                type="datetime"
                placeholder="选择报名截止时间"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="人数上限">
              <el-input-number v-model="form.max_participants" :min="0" placeholder="0表示不限" />
              <span class="form-tip">0表示不限制人数</span>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="活动费用">
              <el-input-number v-model="form.price" :min="0" :precision="2" :step="0.01" />
              <span class="form-tip">0表示免费</span>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="活动详情">
          <RichEditor v-model="form.content" :height="400" />
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import RichEditor from '../../components/RichEditor.vue'
import { activityApi } from '../../api'

const router = useRouter()
const route = useRoute()
const formRef = ref()

const isEdit = computed(() => !!route.params.id)

const form = ref({
  title: '',
  description: '',
  content: '',
  thumbnail: '',
  location: '',
  start_time: '',
  end_time: '',
  reg_start_time: null,
  reg_end_time: null,
  max_participants: 0,
  price: 0,
  status: 0
})

const rules = {
  title: [{ required: true, message: '请输入活动名称', trigger: 'blur' }],
  start_time: [{ required: true, message: '请选择活动开始时间', trigger: 'change' }],
  end_time: [{ required: true, message: '请选择活动结束时间', trigger: 'change' }],
}

const uploadUrl = (import.meta.env.VITE_API_BASE_URL || '') + '/api/admin/upload/image'
const uploadHeaders = computed(() => ({
  Authorization: `Bearer ${localStorage.getItem('token') || ''}`
}))

const handleThumbnailSuccess = (response) => {
  form.value.thumbnail = response.url
}

const loadActivity = async () => {
  if (!isEdit.value) return
  try {
    const data = await activityApi.get(route.params.id)
    form.value = {
      title: data.title,
      description: data.description,
      content: data.content,
      thumbnail: data.thumbnail,
      location: data.location,
      start_time: data.start_time,
      end_time: data.end_time,
      reg_start_time: data.reg_start_time,
      reg_end_time: data.reg_end_time,
      max_participants: data.max_participants,
      price: data.price,
      status: data.status,
    }
  } catch (error) {
    ElMessage.error('加载活动失败')
    router.back()
  }
}

const saveActivity = async (status) => {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  const data = { ...form.value, status }

  try {
    if (isEdit.value) {
      await activityApi.update(route.params.id, data)
      ElMessage.success('更新成功')
    } else {
      await activityApi.create(data)
      ElMessage.success('创建成功')
    }
    router.push('/admin/activities')
  } catch (error) {
    // 错误已在 request.js 拦截器中处理
  }
}

onMounted(() => loadActivity())
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.activity-form {
  max-width: 1000px;
}
.thumbnail-upload {
  display: flex;
  align-items: center;
  gap: 10px;
}
.thumbnail {
  width: 200px;
  height: 150px;
  object-fit: cover;
  border: 1px solid #ddd;
  border-radius: 4px;
}
.upload-icon {
  width: 200px;
  height: 150px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px dashed #d9d9d9;
  border-radius: 4px;
  cursor: pointer;
  font-size: 40px;
  color: #8c939d;
}
.upload-icon:hover {
  border-color: #409eff;
  color: #409eff;
}
.form-tip {
  margin-left: 10px;
  color: #999;
  font-size: 12px;
}
</style>
