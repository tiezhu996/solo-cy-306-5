package router

import (
	"gbevent/internal/constants"
	"gbevent/internal/middleware"

	"github.com/gin-gonic/gin"
)

// registerRegistrationRoutes 报名路由。
func (r *Router) registerRegistrationRoutes(g *gin.RouterGroup) {
	regs := g.Group("/registrations")
	regs.Use(middleware.AuthRequired(r.cfg))
	regs.POST("", r.registration.Create)
	regs.GET("/mine", r.registration.Mine)
	regs.GET("", middleware.RequireRole(constants.RoleOrganizer, constants.RoleAdmin), r.registration.List)
	regs.GET("/export", middleware.RequireRole(constants.RoleOrganizer, constants.RoleAdmin), r.registration.Export)
	regs.POST("/offline", middleware.RequireRole(constants.RoleOrganizer, constants.RoleAdmin), r.registration.OfflineCreate)
	regs.POST("/:id/cancel", r.registration.Cancel)
	regs.POST("/:id/review", middleware.RequireRole(constants.RoleOrganizer, constants.RoleAdmin), r.registration.Review)
}
