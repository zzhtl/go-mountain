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
            <Editor
              v-model="form.content"
              :api-key="tinymceApiKey"
              :init="editorInit"
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
import Editor from '@tinymce/tinymce-vue'
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

// TinyMCE配置
const tinymceApiKey = 'no-api-key' // 使用自托管版本
const uploadUrl = '/api/admin/upload/image'
const uploadHeaders = computed(() => ({
  Authorization: axios.defaults.headers.common['Authorization']
}))

// 富文本编辑器配置
const editorInit = {
  height: 500,
  menubar: true,
  language: 'zh_CN',
  plugins: [
    'advlist', 'autolink', 'lists', 'link', 'image', 'media', 'charmap', 'preview',
    'anchor', 'searchreplace', 'visualblocks', 'code', 'fullscreen',
    'insertdatetime', 'table', 'help', 'wordcount'
  ],
  toolbar: 'undo redo | blocks | ' +
    'bold italic forecolor | alignleft aligncenter ' +
    'alignright alignjustify | bullist numlist outdent indent | ' +
    'image media | removeformat | help',
  content_style: 'body { font-family:Helvetica,Arial,sans-serif; font-size:14px }',
  
  // 图片上传配置
  images_upload_url: uploadUrl,
  images_upload_handler: async (blobInfo, progress) => {
    return new Promise(async (resolve, reject) => {
      const formData = new FormData()
      formData.append('file', blobInfo.blob(), blobInfo.filename())
      
      try {
        const res = await axios.post(uploadUrl, formData, {
          headers: {
            'Content-Type': 'multipart/form-data'
          },
          onUploadProgress: (e) => {
            progress(e.loaded / e.total * 100)
          }
        })
        resolve(res.data.url)
      } catch (error) {
        reject(error.response?.data?.error || '上传失败')
      }
    })
  },
  
  // 视频上传配置
  file_picker_types: 'media',
  file_picker_callback: (callback, value, meta) => {
    if (meta.filetype === 'media') {
      const input = document.createElement('input')
      input.setAttribute('type', 'file')
      input.setAttribute('accept', 'video/*')
      
      input.onchange = async function() {
        const file = this.files[0]
        const formData = new FormData()
        formData.append('file', file)
        
        try {
          const res = await axios.post('/api/admin/upload/video', formData, {
            headers: {
              'Content-Type': 'multipart/form-data'
            }
          })
          callback(res.data.url, { title: file.name })
        } catch (error) {
          ElMessage.error(error.response?.data?.error || '视频上传失败')
        }
      }
      
      input.click()
    }
  }
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

.editor-container {
  width: 100%;
}
</style> 