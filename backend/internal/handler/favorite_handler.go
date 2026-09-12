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

// FavoriteHandler 收藏 HTTP 处理器。
type FavoriteHandler struct {
	svc    *service.FavoriteService
	logger *slog.Logger
}

// NewFavoriteHandler 构造收藏处理器。
func NewFavoriteHandler(svc *service.FavoriteService, logger *slog.Logger) *FavoriteHandler {
	return &FavoriteHandler{svc: svc, logger: logger}
}

// Add 收藏活动。
func (h *FavoriteHandler) Add(c *gin.Context) {
	activityID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Favorite add: invalid activity id")
		return
	}
	f, err := h.svc.Add(middleware.GetUserID(c), activityID)
	if err != nil {
		h.wrapError(c, err, "Favorite add failed")
		return
	}
	OKWithMessage(c, constants.MsgFavoriteAdded, f)
}

// Remove 取消收藏。
func (h *FavoriteHandler) Remove(c *gin.Context) {
	activityID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Favorite remove: invalid activity id")
		return
	}
	if err := h.svc.Remove(middleware.GetUserID(c), activityID); err != nil {
		h.wrapError(c, err, "Favorite remove failed")
		return
	}
	OKWithMessage(c, constants.MsgFavoriteRemoved, nil)
}

// Mine 我的收藏列表。
func (h *FavoriteHandler) Mine(c *gin.Context) {
	var q dto.PageQuery
	_ = c.ShouldBindQuery(&q)
	q.Normalize()
	list, total, err := h.svc.ListMine(middleware.GetUserID(c), q.Page, q.PageSize)
	if err != nil {
		h.wrapError(c, err, "Favorite mine failed")
		return
	}
	OK(c, pageResponse(list, total, q.Page, q.PageSize))
}

func (h *FavoriteHandler) wrapError(c *gin.Context, err error, ctx string) {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		h.logger.Warn("favorite handler error", "context", ctx, "error", appErr.Error())
		Fail(c, appErrorStatus(appErr.Code), appErr.Code, appErr.Message)
		return
	}
	h.logger.Error("favorite handler error", "context", ctx, "error", err.Error())
	Fail(c, http.StatusInternalServerError, constants.CodeInternalError, constants.MsgInternalError)
}
