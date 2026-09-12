package handler

import (
	"net/http"

	"gbevent/internal/constants"

	"github.com/gin-gonic/gin"
)

// OK 统一成功响应。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"code": constants.CodeOK, "message": constants.MsgOK, "data": data})
}

// OKWithMessage 统一成功响应并携带自定义文案。
func OKWithMessage(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, gin.H{"code": constants.CodeOK, "message": message, "data": data})
}

// Fail 直接输出错误响应（供 handler 包装 service 错误后使用）。
func Fail(c *gin.Context, status int, code int, message string) {
	c.AbortWithStatusJSON(status, gin.H{"code": code, "message": message, "data": nil})
}

// pageResponse 分页响应结构。
func pageResponse(list any, total int64, page, pageSize int) gin.H {
	return gin.H{"list": list, "total": total, "page": page, "page_size": pageSize}
}
