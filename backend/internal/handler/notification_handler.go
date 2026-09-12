package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"gbevent/internal/constants"
	"gbevent/internal/dto"
	"gbevent/internal/middleware"
	"gbevent/internal/service"
	"gbevent/internal/util"

	"github.com/gin-gonic/gin"
)

// NotificationHandler 通知 HTTP 处理器。
type NotificationHandler struct {
	svc    *service.NotificationService
	logger *slog.Logger
}

// NewNotificationHandler 构造通知处理器。
func NewNotificationHandler(svc *service.NotificationService, logger *slog.Logger) *NotificationHandler {
	return &NotificationHandler{svc: svc, logger: logger}
}

// Mine 我的通知列表。
func (h *NotificationHandler) Mine(c *gin.Context) {
	var q dto.PageQuery
	_ = c.ShouldBindQuery(&q)
	q.Normalize()
	list, total, err := h.svc.ListMine(middleware.GetUserID(c), q.Page, q.PageSize)
	if err != nil {
		h.wrapError(c, err, "Notification mine failed")
		return
	}
	OK(c, pageResponse(list, total, q.Page, q.PageSize))
}

// MarkRead 标记已读。
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Notification[id] mark read: invalid id")
		return
	}
	if err := h.svc.MarkRead(id, middleware.GetUserID(c)); err != nil {
		h.wrapError(c, err, "Notification mark read failed")
		return
	}
	OK(c, nil)
}

// MarkAllRead 全部已读。
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	if err := h.svc.MarkAllRead(middleware.GetUserID(c)); err != nil {
		h.wrapError(c, err, "Notification mark all read failed")
		return
	}
	OK(c, nil)
}

func (h *NotificationHandler) wrapError(c *gin.Context, err error, ctx string) {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		h.logger.Warn("notification handler error", "context", ctx, "error", appErr.Error())
		Fail(c, appErrorStatus(appErr.Code), appErr.Code, appErr.Message)
		return
	}
	h.logger.Error("notification handler error", "context", ctx, "error", err.Error())
	Fail(c, http.StatusInternalServerError, constants.CodeInternalError, constants.MsgInternalError)
}
