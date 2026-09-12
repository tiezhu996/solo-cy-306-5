package router

import (
	"log/slog"

	"gbevent/internal/config"
	"gbevent/internal/handler"
	"gbevent/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Router 路由装配器。
type Router struct {
	cfg     *config.Config
	db      *gorm.DB
	logger  *slog.Logger
	limiter *middleware.RateLimiter

	user         *handler.UserHandler
	activity     *handler.ActivityHandler
	registration *handler.RegistrationHandler
	checkIn      *handler.CheckInRecordHandler
	comment      *handler.CommentHandler
	favorite     *handler.FavoriteHandler
	notification *handler.NotificationHandler
	upload       *handler.UploadHandler
}

// New 构造路由装配器。
func New(cfg *config.Config, db *gorm.DB, logger *slog.Logger,
	user *handler.UserHandler, activity *handler.ActivityHandler,
	registration *handler.RegistrationHandler, checkIn *handler.CheckInRecordHandler,
	comment *handler.CommentHandler, favorite *handler.FavoriteHandler,
	notification *handler.NotificationHandler, upload *handler.UploadHandler) *Router {
	return &Router{
		cfg: cfg, db: db, logger: logger,
		limiter:      middleware.NewRateLimiter(cfg.RateLimitPerMinute),
		user:         user,
		activity:     activity,
		registration: registration,
		checkIn:      checkIn,
		comment:      comment,
		favorite:     favorite,
		notification: notification,
		upload:       upload,
	}
}

// Setup 装配全部路由并返回引擎。
func (r *Router) Setup() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.RequestID())
	engine.Use(middleware.RequestLogger(r.logger))
	engine.Use(middleware.ErrorHandler(r.logger))
	engine.Use(middleware.CORS(r.cfg))
	engine.Use(middleware.JWTConfig(r.cfg))
	engine.Use(middleware.AuditLog(r.db, r.logger))

	engine.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	engine.GET("/api/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	engine.GET("/api/v1/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	engine.Static("/uploads", r.cfg.UploadDir)

	v1 := engine.Group("/api/v1")
	r.registerAuthRoutes(v1)
	r.registerUserRoutes(v1)
	r.registerActivityRoutes(v1)
	r.registerRegistrationRoutes(v1)
	r.registerCheckInRoutes(v1)
	r.registerCommentRoutes(v1)
	r.registerFavoriteRoutes(v1)
	r.registerNotificationRoutes(v1)
	return engine
}

// auth 登录注册（限流）。
func (r *Router) registerAuthRoutes(g *gin.RouterGroup) {
	auth := g.Group("/auth")
	auth.Use(r.limiter.Limit())
	auth.POST("/register", r.user.Register)
	auth.POST("/login", r.user.Login)
}
