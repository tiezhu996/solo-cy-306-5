package router

import (
	"gbevent/internal/middleware"

	"github.com/gin-gonic/gin"
)

// registerFavoriteRoutes 收藏路由。
func (r *Router) registerFavoriteRoutes(g *gin.RouterGroup) {
	favs := g.Group("/favorites")
	favs.Use(middleware.AuthRequired(r.cfg))
	favs.GET("/mine", r.favorite.Mine)

	act := g.Group("/activities/:id/favorite")
	act.Use(middleware.AuthRequired(r.cfg))
	act.POST("", r.favorite.Add)
	act.DELETE("", r.favorite.Remove)
}
