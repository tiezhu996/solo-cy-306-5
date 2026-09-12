package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gbevent/internal/config"
	"gbevent/internal/handler"
	"gbevent/internal/model"
	"gbevent/internal/repository"
	"gbevent/internal/router"
	"gbevent/internal/service"
	"gbevent/internal/util"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	logger := util.NewLogger(slog.LevelInfo)

	db, err := gorm.Open(mysql.Open(cfg.DBDSN()), &gorm.Config{})
	if err != nil {
		logger.Error("connect database failed", "error", err.Error())
		os.Exit(1)
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.Activity{}, &model.Registration{}, &model.CheckInRecord{},
		&model.Comment{}, &model.Favorite{}, &model.Notification{}, &model.AuditLog{},
	); err != nil {
		logger.Error("auto migrate failed", "error", err.Error())
		os.Exit(1)
	}
	if err := service.NewSeedService(db, logger).Seed(); err != nil {
		logger.Error("seed failed", "error", err.Error())
		os.Exit(1)
	}

	userRepo := repository.NewUserRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	regRepo := repository.NewRegistrationRepository(db)
	checkinRepo := repository.NewCheckInRecordRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	notifyRepo := repository.NewNotificationRepository(db)

	userSvc := service.NewUserService(userRepo, logger)
	activitySvc := service.NewActivityService(activityRepo, regRepo, notifyRepo, checkinRepo, logger)
	regSvc := service.NewRegistrationService(db, regRepo, activitySvc, notifyRepo, logger)
	checkinSvc := service.NewCheckInRecordService(db, checkinRepo, regRepo, activitySvc, notifyRepo, logger)
	commentSvc := service.NewCommentService(commentRepo, activitySvc, logger)
	favoriteSvc := service.NewFavoriteService(favoriteRepo, activitySvc, logger)
	notifySvc := service.NewNotificationService(notifyRepo, logger)

	userHandler := handler.NewUserHandler(userSvc, logger)
	activityHandler := handler.NewActivityHandler(activitySvc, logger)
	regHandler := handler.NewRegistrationHandler(regSvc, logger)
	checkinHandler := handler.NewCheckInRecordHandler(checkinSvc, logger)
	commentHandler := handler.NewCommentHandler(commentSvc, logger)
	favoriteHandler := handler.NewFavoriteHandler(favoriteSvc, logger)
	notifyHandler := handler.NewNotificationHandler(notifySvc, logger)
	uploadHandler := handler.NewUploadHandler(cfg, logger)

	r := router.New(cfg, db, logger, userHandler, activityHandler, regHandler, checkinHandler,
		commentHandler, favoriteHandler, notifyHandler, uploadHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r.Setup(),
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("server starting", "port", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server run failed", "error", err.Error())
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("server shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", "error", err.Error())
	}
	logger.Info("server stopped")
}
