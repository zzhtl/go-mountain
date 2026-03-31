<template>
  <div class="rich-editor">
    <Editor
      v-model="content"
      :init="editorInit"
      :disabled="disabled"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue'
import Editor from '@tinymce/tinymce-vue'
import 'tinymce/tinymce'
import 'tinymce/themes/silver'
import 'tinymce/icons/default'
import 'tinymce/models/dom'

// 插件
import 'tinymce/plugins/advlist'
import 'tinymce/plugins/autolink'
import 'tinymce/plugins/lists'
import 'tinymce/plugins/link'
import 'tinymce/plugins/image'
import 'tinymce/plugins/charmap'
import 'tinymce/plugins/preview'
import 'tinymce/plugins/anchor'
import 'tinymce/plugins/searchreplace'
import 'tinymce/plugins/visualblocks'
import 'tinymce/plugins/code'
import 'tinymce/plugins/fullscreen'
import 'tinymce/plugins/insertdatetime'
import 'tinymce/plugins/media'
import 'tinymce/plugins/table'
import 'tinymce/plugins/wordcount'

import { uploadApi } from '../api'

const props = defineProps({
  modelValue: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
  height: { type: Number, default: 500 }
})

const emit = defineEmits(['update:modelValue'])

const content = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const editorInit = {
  height: props.height,
  language_url: '/tinymce/langs/zh_CN.js',
  language: 'zh_CN',
  skin_url: '/tinymce/skins/ui/oxide',
  content_css: '/tinymce/skins/content/default/content.min.css',
  plugins: [
    'advlist', 'autolink', 'lists', 'link', 'image', 'charmap', 'preview',
    'anchor', 'searchreplace', 'visualblocks', 'code', 'fullscreen',
    'insertdatetime', 'media', 'table', 'wordcount'
  ],
  toolbar: [
    'undo redo | styles | bold italic underline strikethrough | forecolor backcolor',
    'alignleft aligncenter alignright alignjustify | bullist numlist outdent indent | link image media table | code fullscreen preview'
  ].join(' | '),
  menubar: 'file edit view insert format tools table',
  branding: false,
  promotion: false,
  convert_urls: false,
  // 图片上传
  images_upload_handler: async (blobInfo) => {
    const formData = new FormData()
    formData.append('file', blobInfo.blob(), blobInfo.filename())
    const data = await uploadApi.image(formData)
    return data.url
  },
  // 文件选择
  file_picker_types: 'image media',
  file_picker_callback: (callback, value, meta) => {
    const input = document.createElement('input')
    input.setAttribute('type', 'file')

    if (meta.filetype === 'image') {
      input.setAttribute('accept', 'image/*')
    } else if (meta.filetype === 'media') {
      input.setAttribute('accept', 'video/*')
    }

    input.onchange = async () => {
      const file = input.files[0]
      if (!file) return

      const formData = new FormData()
      formData.append('file', file)

      try {
        let data
        if (meta.filetype === 'image') {
          data = await uploadApi.image(formData)
        } else {
          data = await uploadApi.video(formData)
        }
        callback(data.url, { title: file.name })
      } catch (error) {
        console.error('文件上传失败:', error)
      }
    }

    input.click()
  }
}
</script>

<style scoped>
.rich-editor {
  width: 100%;
}

:deep(.tox-tinymce) {
  border-radius: 4px;
}
</style>
