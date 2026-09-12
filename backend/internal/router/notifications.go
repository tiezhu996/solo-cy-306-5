package router

import (
	"gbevent/internal/middleware"

	"github.com/gin-gonic/gin"
)

// registerNotificationRoutes 通知路由。
func (r *Router) registerNotificationRoutes(g *gin.RouterGroup) {
	notifs := g.Group("/notifications")
	notifs.Use(middleware.AuthRequired(r.cfg))
	notifs.GET("/mine", r.notification.Mine)
	notifs.POST("/read-all", r.notification.MarkAllRead)
	notifs.POST("/:id/read", r.notification.MarkRead)
}
