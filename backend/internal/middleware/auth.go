package middleware

import (
	"net/http"
	"strings"

	"gbevent/internal/config"
	"gbevent/internal/constants"
	"gbevent/internal/util"

	"github.com/gin-gonic/gin"
)

const (
	ctxUserID   = "user_id"
	ctxUsername = "username"
	ctxUserRole = "role"
)

// AuthRequired 验证 JWT，将用户信息注入 gin.Context。
func AuthRequired(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": constants.CodeUnauthorized, "message": constants.MsgUnauthorized, "data": nil})
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := util.ParseToken(cfg.JWTSecret, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": constants.CodeUnauthorized, "message": constants.MsgUnauthorized, "data": nil})
			return
		}
		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxUsername, claims.Username)
		c.Set(ctxUserRole, claims.Role)
		c.Next()
	}
}

// GetUserID 从上下文取用户 ID。
func GetUserID(c *gin.Context) uint64 {
	v, _ := c.Get(ctxUserID)
	id, _ := v.(uint64)
	return id
}

// GetUsername 从上下文取用户名。
func GetUsername(c *gin.Context) string {
	v, _ := c.Get(ctxUsername)
	s, _ := v.(string)
	return s
}

// GetUserRole 从上下文取角色。
func GetUserRole(c *gin.Context) string {
	v, _ := c.Get(ctxUserRole)
	s, _ := v.(string)
	return s
}

// JWTConfig 将 JWT 配置注入上下文，供登录/注册处理器使用。
func JWTConfig(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("jwt_secret", cfg.JWTSecret)
		c.Set("jwt_expire_hours", cfg.JWTExpireHours)
		c.Next()
	}
}
