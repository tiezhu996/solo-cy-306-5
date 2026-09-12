package router

import (
	"gbevent/internal/middleware"

	"github.com/gin-gonic/gin"
)

// registerCommentRoutes 评论路由。
func (r *Router) registerCommentRoutes(g *gin.RouterGroup) {
	comments := g.Group("/activities/:id/comments")
	comments.GET("", r.comment.List)
	comments.POST("", middleware.AuthRequired(r.cfg), r.comment.Create)

	mine := g.Group("/comments")
	mine.Use(middleware.AuthRequired(r.cfg))
	mine.GET("/mine", r.comment.Mine)
	mine.DELETE("/:id", r.comment.Delete)
}
