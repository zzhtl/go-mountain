<template>
  <div class="article-edit">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ isEdit ? '编辑文章' : '新增文章' }}</span>
          <div>
            <el-button @click="$router.back()">返回</el-button>
            <el-button type="primary" @click="saveArticle(0)">保存草稿</el-button>
            <el-button type="success" @click="saveArticle(1)">发布文章</el-button>
          </div>
        </div>
      </template>
      
      <el-form :model="form" label-width="100px" class="article-form">
        <el-form-item label="所属栏目" required>
          <el-select v-model="form.column_id" placeholder="请选择栏目">
            <el-option
              v-for="column in columns"
              :key="column.id"
              :label="column.name"
              :value="column.id"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="文章标题" required>
          <el-input v-model="form.title" placeholder="请输入文章标题" />
        </el-form-item>
        
        <el-form-item label="作者">
          <el-input v-model="form.author" placeholder="请输入作者" />
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
        
        <el-form-item label="文章内容">
          <div class="editor-container">
            <QuillEditor
              v-model:content="form.content"
              :toolbar="editorToolbar"
              contentType="html"
              theme="snow"
              @textChange="onEditorTextChange"
              @ready="onEditorReady"
            />
          </div>
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
import { QuillEditor } from '@vueup/vue-quill'
import '@vueup/vue-quill/dist/vue-quill.snow.css'
import axios from 'axios'

const router = useRouter()
const route = useRoute()

const isEdit = computed(() => !!route.params.id)
const columns = ref([])
const form = ref({
  column_id: null,
  title: '',
  author: '',
  thumbnail: '',
  content: '',
  status: 0
})

// 上传相关配置
const uploadUrl = '/api/admin/upload/image'
const uploadVideoUrl = '/api/admin/upload/video'
const uploadHeaders = computed(() => ({
  Authorization: axios.defaults.headers.common['Authorization']
}))

// Quill编辑器工具栏配置
const editorToolbar = [
  ['bold', 'italic', 'underline', 'strike'],
  ['blockquote', 'code-block'],
  [{ 'header': 1 }, { 'header': 2 }],
  [{ 'list': 'ordered' }, { 'list': 'bullet' }],
  [{ 'script': 'sub' }, { 'script': 'super' }],
  [{ 'indent': '-1' }, { 'indent': '+1' }],
  [{ 'direction': 'rtl' }],
  [{ 'size': ['small', false, 'large', 'huge'] }],
  [{ 'header': [1, 2, 3, 4, 5, 6, false] }],
  [{ 'color': [] }, { 'background': [] }],
  [{ 'font': [] }],
  [{ 'align': [] }],
  ['clean'],
  ['link', 'image', 'video']
]

// 编辑器事件处理
const onEditorTextChange = () => {
  // 可以在这里添加内容变化时的处理逻辑
}

const onEditorReady = (quill) => {
  // 自定义图片上传处理
  const toolbar = quill.getModule('toolbar')
  
  // 图片上传处理
  toolbar.addHandler('image', () => {
    const input = document.createElement('input')
    input.setAttribute('type', 'file')
    input.setAttribute('accept', 'image/*')
    input.click()
    
    input.onchange = async () => {
      const file = input.files[0]
      if (file) {
        const formData = new FormData()
        formData.append('file', file)
        
        try {
          ElMessage.info('正在上传图片...')
          const res = await axios.post(uploadUrl, formData, {
            headers: {
              'Content-Type': 'multipart/form-data'
            }
          })
          
          // 获取光标位置并插入图片
          const range = quill.getSelection()
          quill.insertEmbed(range.index, 'image', res.data.url)
          // 光标向后移动一位
          quill.setSelection(range.index + 1)
          ElMessage.success('图片上传成功')
        } catch (error) {
          ElMessage.error('图片上传失败')
        }
      }
    }
  })
  
  // 视频上传处理
  toolbar.addHandler('video', () => {
    const input = document.createElement('input')
    input.setAttribute('type', 'file')
    input.setAttribute('accept', 'video/*')
    input.click()
    
    input.onchange = async () => {
      const file = input.files[0]
      if (file) {
        const formData = new FormData()
        formData.append('file', file)
        
        try {
          ElMessage.info('正在上传视频，请稍候...')
          const res = await axios.post(uploadVideoUrl, formData, {
            headers: {
              'Content-Type': 'multipart/form-data'
            },
            // 添加上传进度监听
            onUploadProgress: (progressEvent) => {
              const percentCompleted = Math.round((progressEvent.loaded * 100) / progressEvent.total)
              if (percentCompleted % 20 === 0) {
                ElMessage.info(`视频上传进度: ${percentCompleted}%`)
              }
            }
          })
          
          // 获取光标位置并插入视频
          const range = quill.getSelection()
          quill.insertEmbed(range.index, 'video', res.data.url)
          // 光标向后移动一位
          quill.setSelection(range.index + 1)
          ElMessage.success('视频上传成功')
        } catch (error) {
          console.error('视频上传失败:', error)
          ElMessage.error(error.response?.data?.error || '视频上传失败')
        }
      }
    }
  })
}

// 加载栏目列表
const loadColumns = async () => {
  try {
    const res = await axios.get('/api/admin/columns/')
    columns.value = res.data
  } catch (error) {
    ElMessage.error('加载栏目失败')
  }
}

// 加载文章详情
const loadArticle = async () => {
  if (!isEdit.value) return
  
  try {
    const res = await axios.get(`/api/admin/articles/${route.params.id}`)
    form.value = {
      column_id: res.data.column_id,
      title: res.data.title,
      author: res.data.author,
      thumbnail: res.data.thumbnail,
      content: res.data.content,
      status: res.data.status
    }
  } catch (error) {
    ElMessage.error('加载文章失败')
    router.back()
  }
}

// 处理缩略图上传成功
const handleThumbnailSuccess = (response) => {
  form.value.thumbnail = response.url
}

// 保存文章
const saveArticle = async (status) => {
  if (!form.value.column_id) {
    ElMessage.warning('请选择栏目')
    return
  }
  
  if (!form.value.title) {
    ElMessage.warning('请输入文章标题')
    return
  }
  
  const data = {
    ...form.value,
    status
  }
  
  try {
    if (isEdit.value) {
      await axios.put(`/api/admin/articles/${route.params.id}`, data)
      ElMessage.success('更新成功')
    } else {
      await axios.post('/api/admin/articles/', data)
      ElMessage.success('创建成功')
    }
    router.push('/admin/articles')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '保存失败')
  }
}

onMounted(() => {
  loadColumns()
  loadArticle()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.article-form {
  max-width: 1200px;
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

/* 修复卡片和表单样式，使其可以扩展 */
:deep(.el-card) {
  height: auto;
  overflow: visible;
}

:deep(.el-card__body) {
  height: auto;
  overflow: visible;
  padding-bottom: 50px;
}

:deep(.el-form-item) {
  margin-bottom: 20px;
}

:deep(.el-form-item__content) {
  height: auto;
  overflow: visible;
}

/* 编辑器容器样式 */
.editor-container {
  width: 100%;
  min-height: 500px;
  height: auto;
  position: relative;
}

/* Quill编辑器样式 */
:deep(.quill-editor) {
  height: auto !important;
  min-height: 500px;
  max-height: none !important;
}

:deep(.ql-container) {
  height: auto !important;
  min-height: 500px;
  max-height: none !important;
  overflow: visible;
}

:deep(.ql-editor) {
  min-height: 500px;
  height: auto !important;
  max-height: none !important;
  overflow-y: visible;
}

/* 工具栏固定在顶部 */
:deep(.ql-toolbar) {
  position: sticky;
  top: 0;
  z-index: 10;
  background-color: white;
  border-bottom: 1px solid #ccc;
}

/* 视频样式修复 */
:deep(.ql-editor .ql-video) {
  display: block;
  max-width: 100%;
  width: 640px; /* 设置合适的视频宽度 */
  height: 360px; /* 设置16:9比例的高度 */
  margin: 15px auto;
}

/* 确保所有媒体元素不超出容器 */
:deep(.ql-editor img),
:deep(.ql-editor video) {
  max-width: 100%;
  height: auto;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}
</style> 