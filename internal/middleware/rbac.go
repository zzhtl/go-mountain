package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/zzhtl/go-mountain/internal/model"
	"github.com/zzhtl/go-mountain/internal/pkg/response"
)

// RBACAuth 基于角色的 API 权限校验中间件
// 工作原理：
//  1. 从 JWT claims 中获取 role_id
//  2. 如果是 admin 角色，直接放行
//  3. 根据请求路径和方法，匹配 menus 表中 type=3 的权限记录
//  4. 查询 role_menus 关联表判断角色是否拥有该权限
func RBACAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取角色信息
		roleName, exists := c.Get("role")
		if !exists {
			response.Forbidden(c, "无权限访问")
			c.Abort()
			return
		}

		// admin 角色直接放行
		if roleName == "admin" {
			c.Next()
			return
		}

		// 获取 role_id
		roleIDVal, exists := c.Get("role_id")
		if !exists {
			response.Forbidden(c, "无权限访问")
			c.Abort()
			return
		}
		roleID := int64(roleIDVal.(float64))

		// 从请求路径和方法解析权限标识
		permission := resolvePermission(c.FullPath(), c.Request.Method)
		if permission == "" {
			// 无法解析权限标识时放行（兼容尚未配置权限的路由）
			c.Next()
			return
		}

		// 查询角色是否拥有该 API 权限
		var count int64
		db.Model(&model.RoleMenu{}).
			Joins("INNER JOIN menus ON menus.id = role_menus.menu_id").
			Where("role_menus.role_id = ? AND menus.permission = ? AND menus.type = 3 AND menus.status = 1",
				roleID, permission).
			Count(&count)

		if count == 0 {
			response.Forbidden(c, "无权限访问")
			c.Abort()
			return
		}

		c.Next()
	}
}

// resolvePermission 从路由路径和HTTP方法解析权限标识
// 路径格式: /api/admin/{module}/... → 权限: {module}:{action}
// 示例:
//
//	GET  /api/admin/articles/       → article:list
//	POST /api/admin/articles/       → article:create
//	GET  /api/admin/articles/:id    → article:get
//	PUT  /api/admin/articles/:id    → article:update
//	DELETE /api/admin/articles/:id  → article:delete
func resolvePermission(path, method string) string {
	// 去掉 /api/admin/ 前缀
	path = strings.TrimPrefix(path, "/api/admin/")
	if path == "" {
		return ""
	}

	// 拆分路径段
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}

	// 第一段是模块名（去掉复数后缀作为单数形式）
	module := parts[0]
	// 处理不规则复数：activities → activity
	if strings.HasSuffix(module, "ies") {
		module = module[:len(module)-3] + "y"
	} else {
		module = strings.TrimSuffix(module, "s")
	}
	// 特殊处理：backend-users → backend_user
	module = strings.ReplaceAll(module, "-", "_")

	// 根据路径结构和 HTTP 方法确定操作
	action := resolveAction(parts, method)

	return module + ":" + action
}

// resolveAction 根据路径段和 HTTP 方法确定操作名
func resolveAction(parts []string, method string) string {
	hasID := len(parts) >= 2 && strings.HasPrefix(parts[1], ":")

	// 处理特殊子路由，如 /:id/status、/:id/menus、/current/menus
	if len(parts) >= 3 {
		subAction := parts[len(parts)-1]
		switch method {
		case "GET":
			return subAction
		case "PUT", "POST":
			return "update_" + subAction
		}
	}

	switch method {
	case "GET":
		if hasID {
			return "get"
		}
		return "list"
	case "POST":
		return "create"
	case "PUT":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return "unknown"
	}
}
