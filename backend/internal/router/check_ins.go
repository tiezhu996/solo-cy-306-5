package router

import (
	"gbevent/internal/constants"
	"gbevent/internal/middleware"

	"github.com/gin-gonic/gin"
)

// registerCheckInRoutes 签到路由。
func (r *Router) registerCheckInRoutes(g *gin.RouterGroup) {
	checkins := g.Group("/check-ins")
	checkins.Use(middleware.AuthRequired(r.cfg), middleware.RequireRole(constants.RoleOrganizer, constants.RoleAdmin))
	checkins.POST("", r.checkIn.CheckIn)
	checkins.GET("", r.checkIn.Records)
}
