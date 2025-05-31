// 简化的图片处理工具函数
export const createImageUploadHandler = (uploadHandler) => {
  return {
    // 安全的图片插入方法
    safeInsertImage: (quill, index, url) => {
      try {
        quill.enable()
        
        setTimeout(() => {
          try {
            const range = quill.getSelection() || { index: index, length: 0 }
            quill.insertText(range.index, '\n', 'user')
            
            setTimeout(() => {
              try {
                quill.insertEmbed(range.index, 'image', url, 'user')
                quill.setSelection(range.index + 1, 0, 'user')
              } catch (embedError) {
                console.error('图片插入失败:', embedError)
              }
            }, 50)
          } catch (insertError) {
            console.error('图片插入过程失败:', insertError)
          }
        }, 100)
      } catch (error) {
        console.error('图片插入准备失败:', error)
      }
    },

    // 处理粘贴图片
    handlePaste: (e, quill) => {
      const items = e.clipboardData?.items
      if (!items) return false
      
      for (let i = 0; i < items.length; i++) {
        const item = items[i]
        if (item.type.indexOf('image') !== -1) {
          e.preventDefault()
          const file = item.getAsFile()
          if (uploadHandler) {
            uploadHandler(file).then(url => {
              const range = quill.getSelection() || { index: quill.getLength(), length: 0 }
              
              // 直接调用安全插入方法
              try {
                quill.enable()
                setTimeout(() => {
                  try {
                    quill.insertText(range.index, '\n', 'user')
                    setTimeout(() => {
                      try {
                        quill.insertEmbed(range.index, 'image', url, 'user')
                        quill.setSelection(range.index + 1, 0, 'user')
                      } catch (embedError) {
                        console.error('图片插入失败:', embedError)
                      }
                    }, 50)
                  } catch (insertError) {
                    console.error('图片插入过程失败:', insertError)
                  }
                }, 100)
              } catch (error) {
                console.error('图片插入准备失败:', error)
              }
            }).catch(error => {
              console.error('粘贴图片上传失败:', error)
            })
          }
          return true
        }
      }
      return false
    },

    // 处理拖拽图片
    handleDrop: (e, quill) => {
      e.preventDefault()
      const files = e.dataTransfer.files
      
      for (let i = 0; i < files.length; i++) {
        const file = files[i]
        if (file.type.indexOf('image') !== -1) {
          if (uploadHandler) {
            uploadHandler(file).then(url => {
              const range = quill.getSelection() || { index: quill.getLength(), length: 0 }
              
              // 直接调用安全插入方法
              try {
                quill.enable()
                setTimeout(() => {
                  try {
                    quill.insertText(range.index, '\n', 'user')
                    setTimeout(() => {
                      try {
                        quill.insertEmbed(range.index, 'image', url, 'user')
                        quill.setSelection(range.index + 1, 0, 'user')
                      } catch (embedError) {
                        console.error('图片插入失败:', embedError)
                      }
                    }, 50)
                  } catch (insertError) {
                    console.error('图片插入过程失败:', insertError)
                  }
                }, 100)
              } catch (error) {
                console.error('图片插入准备失败:', error)
              }
            }).catch(error => {
              console.error('拖拽图片上传失败:', error)
            })
          }
          break
        }
      }
    },

    // 处理拖拽悬停
    handleDragOver: (e) => {
      e.preventDefault()
    }
  }
}

export default createImageUploadHandler 