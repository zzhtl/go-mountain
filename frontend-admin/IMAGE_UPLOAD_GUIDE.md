# 富文本编辑器图片上传与调整功能说明

## 功能概述

本系统的富文本编辑器已经升级，解决了原有的 `HierarchyRequestError` 问题，并新增了图片拖拽缩放功能。

## 主要改进

### 1. 解决 HierarchyRequestError 问题

**问题原因**: 
- Quill 编辑器在插入图片时，由于 DOM 操作时序问题导致父子元素层级冲突

**解决方案**:
- 使用安全的图片插入方法 `safeInsertImage`
- 通过 `setTimeout` 和 `nextTick` 确保 DOM 操作在正确的时序中执行
- 添加错误处理机制，防止插入失败

### 2. 新增图片调整功能

**功能特性**:
- 双击图片打开调整对话框
- 支持宽度和高度精确调整
- 支持保持比例缩放
- 支持对齐方式设置（左对齐/居中/右对齐）
- 实时预览调整效果

## 使用方法

### 图片上传方式

1. **工具栏上传**: 点击编辑器工具栏的图片按钮
2. **拖拽上传**: 直接将图片拖拽到编辑器内容区域
3. **粘贴上传**: 复制图片后在编辑器中按 Ctrl+V 粘贴

### 图片调整方法

1. **打开调整对话框**: 双击编辑器中的任意图片
2. **调整尺寸**: 在对话框中修改宽度和高度数值
3. **保持比例**: 勾选"保持比例"选项，修改宽度时自动计算高度
4. **设置对齐**: 选择左对齐、居中或右对齐
5. **确认应用**: 点击"确定"按钮应用调整

### 图片规格限制

- **支持格式**: JPG, JPEG, PNG, GIF, WebP
- **文件大小**: 最大 5MB
- **推荐尺寸**: 宽度不超过 800px，以保证最佳显示效果

## 技术实现

### 核心组件

1. **ArticleEdit.vue**: 主编辑器组件
2. **quill-image-resize.js**: 自定义图片处理模块

### 关键功能

```javascript
// 安全的图片插入方法
const safeInsertImage = (quill, index, url) => {
  try {
    quill.enable()
    setTimeout(() => {
      try {
        const range = quill.getSelection() || { index: index, length: 0 }
        quill.insertText(range.index, '\n', 'user')
        nextTick(() => {
          try {
            quill.insertEmbed(range.index, 'image', url, 'user')
            quill.setSelection(range.index + 1, 0, 'user')
          } catch (embedError) {
            console.error('图片插入失败:', embedError)
            ElMessage.error('图片插入失败，请重试')
          }
        })
      } catch (insertError) {
        console.error('图片插入过程失败:', insertError)
        ElMessage.error('图片插入失败，请重试')
      }
    }, 100)
  } catch (error) {
    console.error('图片插入准备失败:', error)
    ElMessage.error('图片插入失败，请重试')
  }
}
```

### 图片调整配置

```javascript
const imageConfig = ref({
  width: 400,        // 图片宽度
  height: 300,       // 图片高度
  align: 'center',   // 对齐方式: left/center/right
  keepRatio: true,   // 是否保持比例
  originalRatio: 1   // 原始宽高比
})
```

## 样式优化

### 编辑器样式

- 图片最大宽度限制为 100%，防止溢出
- 添加悬停缩放效果，提升用户体验
- 设置圆角和阴影，美化显示效果

### 调整对话框样式

- 左侧预览区域，右侧控制面板
- 响应式布局，适配不同屏幕尺寸
- 直观的控件设计，易于操作

## 浏览器兼容性

- 支持现代浏览器 (Chrome 70+, Firefox 65+, Safari 12+)
- 支持移动端浏览器
- 不支持 IE 浏览器

## 问题排查

### 常见问题

1. **图片上传失败**
   - 检查文件格式和大小是否符合要求
   - 确认网络连接正常
   - 查看浏览器控制台错误信息

2. **图片调整无效**
   - 确保双击的是图片元素
   - 检查图片是否加载完成
   - 刷新页面重试

3. **编辑器卡顿**
   - 避免上传过大的图片文件
   - 清理浏览器缓存
   - 检查内存使用情况

### 调试方法

1. 打开浏览器开发者工具
2. 查看 Console 面板的错误信息
3. 检查 Network 面板的网络请求
4. 查看 Elements 面板的 DOM 结构

## 后续优化计划

1. 支持图片裁剪功能
2. 添加图片压缩选项
3. 支持图片水印添加
4. 优化大图片加载性能
5. 添加图片格式转换功能

## 更新记录

- **v1.0.0**: 解决 HierarchyRequestError 问题
- **v1.1.0**: 新增图片双击调整功能
- **v1.2.0**: 添加拖拽和粘贴上传支持
- **v1.3.0**: 优化样式和用户体验 