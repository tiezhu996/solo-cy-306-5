package router

import (
	"gbevent/internal/constants"
	"gbevent/internal/middleware"

	"github.com/gin-gonic/gin"
)

// registerActivityRoutes 活动路由。
func (r *Router) registerActivityRoutes(g *gin.RouterGroup) {
	activities := g.Group("/activities")
	activities.GET("", r.activity.List)
	activities.GET("/calendar", r.activity.Calendar)
	activities.GET("/mine", middleware.AuthRequired(r.cfg), middleware.RequireRole(constants.RoleOrganizer, constants.RoleAdmin), r.activity.Mine)
	activities.GET("/:id", r.activity.Get)
	activities.GET("/:id/stats", middleware.AuthRequired(r.cfg), middleware.RequireRole(constants.RoleOrganizer, constants.RoleAdmin), r.activity.Stats)

	auth := activities.Group("")
	auth.Use(middleware.AuthRequired(r.cfg), middleware.RequireRole(constants.RoleOrganizer, constants.RoleAdmin))
	auth.POST("", r.activity.Create)
	auth.PUT("/:id", r.activity.Update)
	auth.POST("/:id/publish", r.activity.Publish)
	auth.POST("/:id/end", r.activity.End)
	auth.DELETE("/:id", r.activity.Delete)
}
