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
          <div class="editor-tip">
            <el-alert
              title="图片调整提示"
              type="info"
              :closable="false"
              show-icon
            >
              <template #default>
                <div>上传图片后，可以通过以下方式调整图片尺寸：</div>
                <ul style="margin: 5px 0; padding-left: 20px;">
                  <li>双击图片打开调整对话框</li>
                  <li>右键点击图片选择调整</li>
                  <li>选中图片后按Enter键</li>
                </ul>
              </template>
            </el-alert>
          </div>
          <div class="editor-container">
            <QuillEditor
              ref="editorRef"
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
    
    <!-- 图片调整对话框 -->
    <el-dialog v-model="showImageDialog" title="调整图片" width="600px">
      <div class="image-adjust-container">
        <div class="image-preview">
          <img ref="adjustImageRef" :src="currentImageSrc" :style="imageStyle" />
        </div>
        <div class="image-controls">
          <el-form label-width="80px">
            <el-form-item label="宽度">
              <el-input-number
                v-model="imageConfig.width"
                :min="100"
                :max="800"
                @change="updateImageStyle"
              />
            </el-form-item>
            <el-form-item label="高度">
              <el-input-number
                v-model="imageConfig.height"
                :min="50"
                :max="600"
                @change="updateImageStyle"
              />
            </el-form-item>
            <el-form-item label="对齐方式">
              <el-radio-group v-model="imageConfig.align" @change="updateImageStyle">
                <el-radio label="left">左对齐</el-radio>
                <el-radio label="center">居中</el-radio>
                <el-radio label="right">右对齐</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item>
              <el-checkbox v-model="imageConfig.keepRatio" @change="toggleKeepRatio">
                保持比例
              </el-checkbox>
            </el-form-item>
          </el-form>
        </div>
      </div>
      <template #footer>
        <el-button @click="showImageDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmImageAdjust">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { QuillEditor } from '@vueup/vue-quill'
import '@vueup/vue-quill/dist/vue-quill.snow.css'
import axios from 'axios'
import { createImageUploadHandler } from '../utils/quill-image-resize.js'

const router = useRouter()
const route = useRoute()
const editorRef = ref(null)

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

// 图片调整相关
const showImageDialog = ref(false)
const currentImageSrc = ref('')
const currentImageBlot = ref(null)
const adjustImageRef = ref(null)
const imageConfig = ref({
  width: 400,
  height: 300,
  align: 'center',
  keepRatio: true,
  originalRatio: 1
})

// 上传相关配置
const uploadUrl = '/api/admin/upload/image'
const uploadVideoUrl = '/api/admin/upload/video'
const uploadHeaders = computed(() => ({
  Authorization: axios.defaults.headers.common['Authorization']
}))

// 图片上传处理函数
const uploadImageHandler = async (file) => {
  const formData = new FormData()
  formData.append('file', file)
  
  try {
    ElMessage.info('正在上传图片...')
    const res = await axios.post(uploadUrl, formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
    ElMessage.success('图片上传成功')
    return res.data.url
  } catch (error) {
    console.error('图片上传失败:', error)
    ElMessage.error('图片上传失败')
    throw error
  }
}

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

// 创建图片处理器
const imageHandler = createImageUploadHandler(uploadImageHandler)

// 计算图片样式 - 只用于预览，不影响实际对齐
const imageStyle = computed(() => ({
  width: `${imageConfig.value.width}px`,
  height: `${imageConfig.value.height}px`,
  display: 'block',
  maxWidth: '100%',
  objectFit: 'contain'
}))

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
        try {
          const url = await uploadImageHandler(file)
          const range = quill.getSelection()
          if (range) {
            imageHandler.safeInsertImage(quill, range.index, url)
          } else {
            imageHandler.safeInsertImage(quill, quill.getLength(), url)
          }
        } catch (error) {
          console.error('图片上传失败:', error)
        }
      }
    }
  })
  
  // 添加图片双击事件监听，实现图片调整功能
  const editorElement = quill.root
  editorElement.addEventListener('dblclick', (e) => {
    console.log('双击事件触发，目标元素:', e.target.tagName, e.target)
    if (e.target.tagName === 'IMG') {
      console.log('检测到图片双击，打开调整对话框')
      e.preventDefault()
      e.stopPropagation()
      openImageAdjustDialog(e.target, quill)
    }
  })
  
  // 添加单击事件作为备选方案，长按或右键点击
  editorElement.addEventListener('contextmenu', (e) => {
    if (e.target.tagName === 'IMG') {
      e.preventDefault()
      openImageAdjustDialog(e.target, quill)
    }
  })
  
  // 添加键盘事件，选中图片后按Enter键调整
  editorElement.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' && e.target.tagName === 'IMG') {
      e.preventDefault()
      openImageAdjustDialog(e.target, quill)
    }
  })
  
  // 添加粘贴图片处理
  editorElement.addEventListener('paste', (e) => {
    imageHandler.handlePaste(e, quill)
  })
  
  // 添加拖拽图片处理
  editorElement.addEventListener('drop', (e) => {
    imageHandler.handleDrop(e, quill)
  })
  
  editorElement.addEventListener('dragover', (e) => {
    imageHandler.handleDragOver(e)
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

// 打开图片调整对话框
const openImageAdjustDialog = (imgElement, quill) => {
  console.log('打开图片调整对话框，图片元素:', imgElement)
  console.log('图片src:', imgElement.src)
  console.log('图片尺寸:', imgElement.width, 'x', imgElement.height)
  
  currentImageSrc.value = imgElement.src
  currentImageBlot.value = { element: imgElement, quill }
  
  // 获取当前图片尺寸
  const rect = imgElement.getBoundingClientRect()
  const computedStyle = window.getComputedStyle(imgElement)
  
  // 优先使用元素的显示尺寸
  let width = parseInt(computedStyle.width) || imgElement.offsetWidth || imgElement.width || rect.width || 400
  let height = parseInt(computedStyle.height) || imgElement.offsetHeight || imgElement.height || rect.height || 300
  
  imageConfig.value.width = Math.round(width)
  imageConfig.value.height = Math.round(height)
  
  console.log('设置的宽度高度:', imageConfig.value.width, 'x', imageConfig.value.height)
  
  // 计算原始比例
  if (imgElement.naturalWidth && imgElement.naturalHeight) {
    imageConfig.value.originalRatio = imgElement.naturalWidth / imgElement.naturalHeight
  } else {
    imageConfig.value.originalRatio = imageConfig.value.width / imageConfig.value.height
  }
  
  console.log('原始比例:', imageConfig.value.originalRatio)
  
  // 获取对齐方式：优先检查类名，然后检查样式
  console.log('图片样式信息:', {
    className: imgElement.className,
    dataAlign: imgElement.getAttribute('data-align'),
    styleFloat: imgElement.style.float,
    computedFloat: computedStyle.float,
    styleMargin: imgElement.style.margin,
    styleMarginLeft: imgElement.style.marginLeft,
    styleMarginRight: imgElement.style.marginRight,
    computedMarginLeft: computedStyle.marginLeft,
    computedMarginRight: computedStyle.marginRight
  })
  
  // 优先根据类名判断对齐方式
  if (imgElement.classList.contains('image-center')) {
    imageConfig.value.align = 'center'
  } else if (imgElement.classList.contains('image-right')) {
    imageConfig.value.align = 'right'
  } else if (imgElement.classList.contains('image-left')) {
    imageConfig.value.align = 'left'
  } else if (imgElement.getAttribute('data-align')) {
    // 根据data属性判断
    imageConfig.value.align = imgElement.getAttribute('data-align')
  } else if (computedStyle.float === 'right' || imgElement.style.float === 'right') {
    imageConfig.value.align = 'right'
  } else if (computedStyle.float === 'left' || imgElement.style.float === 'left') {
    imageConfig.value.align = 'left'
  } else if ((computedStyle.marginLeft === 'auto' && computedStyle.marginRight === 'auto') ||
             (imgElement.style.marginLeft === 'auto' && imgElement.style.marginRight === 'auto')) {
    imageConfig.value.align = 'center'
  } else {
    // 默认为左对齐
    imageConfig.value.align = 'left'
  }
  
  console.log('检测到的对齐方式:', imageConfig.value.align)
  
  showImageDialog.value = true
  ElMessage.info('双击图片可调整尺寸，右键图片也可以调整')
}

// 更新图片样式预览
const updateImageStyle = () => {
  if (imageConfig.value.keepRatio) {
    // 保持比例时，根据宽度自动计算高度
    imageConfig.value.height = Math.round(imageConfig.value.width / imageConfig.value.originalRatio)
  }
}

// 切换保持比例
const toggleKeepRatio = () => {
  if (imageConfig.value.keepRatio) {
    updateImageStyle()
  }
}

// 确认图片调整
const confirmImageAdjust = () => {
  if (currentImageBlot.value) {
    const { element, quill } = currentImageBlot.value
    
    console.log('应用图片调整，新尺寸:', imageConfig.value.width, 'x', imageConfig.value.height)
    console.log('对齐方式:', imageConfig.value.align)
    console.log('图片当前DOM结构:', element.parentElement)
    
    // 设置图片基础样式
    element.style.width = `${imageConfig.value.width}px`
    element.style.height = `${imageConfig.value.height}px`
    element.style.maxWidth = 'none'
    element.style.objectFit = 'contain'
    element.setAttribute('width', imageConfig.value.width)
    element.setAttribute('height', imageConfig.value.height)
    
    // 彻底清除所有现有的对齐样式和定位样式
    element.style.margin = ''
    element.style.marginLeft = ''
    element.style.marginRight = ''
    element.style.marginTop = ''
    element.style.marginBottom = ''
    element.style.float = ''
    element.style.cssFloat = ''
    element.style.styleFloat = ''
    element.style.display = ''
    element.style.position = ''
    element.style.left = ''
    element.style.right = ''
    element.style.top = ''
    element.style.bottom = ''
    element.style.transform = ''
    element.style.textAlign = ''
    element.className = '' // 清除所有类名
    element.removeAttribute('data-align') // 清除对齐属性
    
    // 处理父元素样式
    const parent = element.parentElement
    if (parent) {
      parent.style.textAlign = ''
      parent.style.margin = ''
      parent.style.marginLeft = ''
      parent.style.marginRight = ''
      parent.style.display = ''
      parent.style.float = ''
      parent.style.cssFloat = ''
      parent.style.styleFloat = ''
      parent.className = parent.className.replace(/ql-align-\w+/g, '')
    }
    
    // 彻底清除所有样式，确保对齐切换正常
    console.log(`应用 ${imageConfig.value.align} 对齐`)
    
    // 1. 清除所有对齐类名
    element.className = element.className.replace(/image-(left|center|right)/g, '').trim()
    
    // 2. 清除所有可能影响对齐的内联样式
    element.style.float = ''
    element.style.cssFloat = ''
    element.style.styleFloat = ''
    element.style.margin = ''
    element.style.marginLeft = ''
    element.style.marginRight = ''
    element.style.display = ''
    element.style.position = ''
    element.style.left = ''
    element.style.right = ''
    element.style.transform = ''
    
    // 3. 强制刷新DOM
    element.offsetHeight // 触发重排
    
    // 4. 应用新的对齐方式
    if (imageConfig.value.align === 'center') {
      element.className += ' image-center'
      console.log('✅ 设置为居中')
    } else if (imageConfig.value.align === 'right') {
      element.className += ' image-right'
      console.log('✅ 设置为右对齐')
    } else {
      element.className += ' image-left'
      console.log('✅ 设置为左对齐')
    }
    
    // 5. 记录对齐方式
    element.setAttribute('data-align', imageConfig.value.align)
    
    // 存储对齐信息到data属性
    element.setAttribute('data-align', imageConfig.value.align)
    element.setAttribute('data-custom-size', 'true')
    
    // 强制刷新编辑器
    setTimeout(() => {
      if (quill) {
        quill.update('user')
        
        // 验证样式应用结果
        const finalStyles = window.getComputedStyle(element)
        console.log('最终计算样式:', {
          width: finalStyles.width,
          height: finalStyles.height,
          margin: finalStyles.margin,
          marginLeft: finalStyles.marginLeft,
          marginRight: finalStyles.marginRight,
          display: finalStyles.display,
          float: finalStyles.float,
          textAlign: parent ? window.getComputedStyle(parent).textAlign : 'none'
        })
      }
    }, 100)
    
    showImageDialog.value = false
    ElMessage.success(`图片调整成功！新尺寸: ${imageConfig.value.width}x${imageConfig.value.height}，对齐: ${imageConfig.value.align}`)
  } else {
    ElMessage.error('未找到要调整的图片元素')
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
  width: 640px;
  height: 360px;
  margin: 15px auto;
}

/* 基础图片样式 */
:deep(.ql-editor img) {
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: transform 0.2s ease;
}

:deep(.ql-editor img:hover) {
  transform: scale(1.02);
  outline: 2px solid #409eff;
}

/* 图片被选中时的样式 */
:deep(.ql-editor img:focus) {
  outline: 3px solid #409eff;
  outline-offset: 2px;
}

/* 自定义尺寸的图片基础样式 */
:deep(.ql-editor img[data-custom-size="true"]) {
  max-width: none !important;
  height: auto !important;
}

/* 左对齐图片 - 强制覆盖所有其他样式 */
:deep(.ql-editor img.image-left) {
  display: block !important;
  float: left !important;
  margin: 10px 10px 10px 0 !important;
  position: static !important;
  left: auto !important;
  right: auto !important;
  transform: none !important;
}

/* 右对齐图片 - 强制覆盖所有其他样式 */
:deep(.ql-editor img.image-right) {
  display: block !important;
  float: right !important;
  margin: 10px 0 10px 10px !important;
  position: static !important;
  left: auto !important;
  right: auto !important;
  transform: none !important;
}

/* 居中图片 - 强制覆盖所有其他样式 */
:deep(.ql-editor img.image-center) {
  display: block !important;
  float: none !important;
  margin: 10px auto !important;
  position: static !important;
  left: auto !important;
  right: auto !important;
  transform: none !important;
}

/* 通用图片容器样式 */
:deep(.ql-editor p) {
  margin: 10px 0;
  min-height: 1em;
}

/* 清除浮动，防止布局问题 */
:deep(.ql-editor p:after) {
  content: "";
  display: table;
  clear: both;
}



/* 图片调整提示样式 */
.editor-tip {
  margin-bottom: 10px;
}

.editor-tip .el-alert {
  border-radius: 6px;
}

:deep(.ql-editor video) {
  max-width: 100%;
  height: auto;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

/* 图片调整对话框样式 */
.image-adjust-container {
  display: flex;
  gap: 20px;
}

.image-preview {
  flex: 1;
  text-align: center;
  border: 1px solid #ddd;
  border-radius: 4px;
  padding: 20px;
  background-color: #f9f9f9;
}

.image-preview img {
  max-width: 100%;
  max-height: 300px;
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.image-controls {
  width: 200px;
}
</style> 