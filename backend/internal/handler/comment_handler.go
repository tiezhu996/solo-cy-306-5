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

// CommentHandler 评论 HTTP 处理器。
type CommentHandler struct {
	svc    *service.CommentService
	logger *slog.Logger
}

// NewCommentHandler 构造评论处理器。
func NewCommentHandler(svc *service.CommentService, logger *slog.Logger) *CommentHandler {
	return &CommentHandler{svc: svc, logger: logger}
}

// List 活动评论列表。
func (h *CommentHandler) List(c *gin.Context) {
	activityID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Comment list: invalid activity id")
		return
	}
	list, err := h.svc.ListByActivity(activityID)
	if err != nil {
		h.wrapError(c, err, "Comment list failed")
		return
	}
	avg, _ := h.svc.AvgRating(activityID)
	OK(c, gin.H{"list": list, "avg_rating": avg})
}

// Create 发表评论。
func (h *CommentHandler) Create(c *gin.Context) {
	activityID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Comment create: invalid activity id")
		return
	}
	var req dto.CommentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Comment[activity_id="+strconv.FormatUint(activityID, 10)+"] create: "+err.Error())
		return
	}
	cm, err := h.svc.Create(activityID, middleware.GetUserID(c), req.Rating, req.Content)
	if err != nil {
		h.wrapError(c, err, "Comment create failed")
		return
	}
	OKWithMessage(c, constants.MsgCommentSuccess, cm)
}

// Mine 我的评论。
func (h *CommentHandler) Mine(c *gin.Context) {
	list, err := h.svc.ListMine(middleware.GetUserID(c))
	if err != nil {
		h.wrapError(c, err, "Comment mine failed")
		return
	}
	OK(c, list)
}

// Delete 删除评论。
func (h *CommentHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Comment[id] delete: invalid id")
		return
	}
	if err := h.svc.Delete(id, middleware.GetUserID(c), middleware.GetUserRole(c)); err != nil {
		h.wrapError(c, err, "Comment delete failed")
		return
	}
	OK(c, nil)
}

func (h *CommentHandler) wrapError(c *gin.Context, err error, ctx string) {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		h.logger.Warn("comment handler error", "context", ctx, "error", appErr.Error())
		Fail(c, appErrorStatus(appErr.Code), appErr.Code, appErr.Message)
		return
	}
	h.logger.Error("comment handler error", "context", ctx, "error", err.Error())
	Fail(c, http.StatusInternalServerError, constants.CodeInternalError, constants.MsgInternalError)
}
