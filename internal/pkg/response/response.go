package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一 API 响应结构
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// PageResult 分页结果
type PageResult struct {
	List     any   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// OK 成功响应
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Created 创建成功响应
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// PageOK 分页成功响应
func PageOK(c *gin.Context, list any, total int64, page, pageSize int) {
	OK(c, PageResult{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// Fail 失败响应
func Fail(c *gin.Context, httpCode int, code int, message string) {
	c.JSON(httpCode, Response{
		Code:    code,
		Message: message,
	})
}

// BadRequest 参数错误
func BadRequest(c *gin.Context, message string) {
	Fail(c, http.StatusBadRequest, 400, message)
}

// Unauthorized 未授权
func Unauthorized(c *gin.Context, message string) {
	Fail(c, http.StatusUnauthorized, 401, message)
}

// Forbidden 无权限
func Forbidden(c *gin.Context, message string) {
	Fail(c, http.StatusForbidden, 403, message)
}

// NotFound 资源不存在
func NotFound(c *gin.Context, message string) {
	Fail(c, http.StatusNotFound, 404, message)
}

// ServerError 服务器内部错误
func ServerError(c *gin.Context, message string) {
	Fail(c, http.StatusInternalServerError, 500, message)
}

// NoContent 无内容响应（删除成功）
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
