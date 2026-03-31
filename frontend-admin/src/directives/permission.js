import { usePermissionStore } from '../store/permission'

// v-permission 指令：按钮级别权限控制
// 用法: v-permission="'article:delete'" 或 v-permission="['article:delete', 'article:update']"
export const vPermission = {
  mounted(el, binding) {
    const permissionStore = usePermissionStore()
    const value = binding.value

    if (!value) return

    const perms = Array.isArray(value) ? value : [value]
    const hasAny = perms.some(p => permissionStore.hasPermission(p))

    if (!hasAny) {
      el.parentNode?.removeChild(el)
    }
  }
}

export default {
  install(app) {
    app.directive('permission', vPermission)
  }
}
