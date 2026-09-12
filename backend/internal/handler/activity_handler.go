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

// ActivityHandler 活动 HTTP 处理器。
type ActivityHandler struct {
	svc    *service.ActivityService
	logger *slog.Logger
}

// NewActivityHandler 构造活动处理器。
func NewActivityHandler(svc *service.ActivityService, logger *slog.Logger) *ActivityHandler {
	return &ActivityHandler{svc: svc, logger: logger}
}

// List 活动列表。
func (h *ActivityHandler) List(c *gin.Context) {
	var q dto.PageQuery
	_ = c.ShouldBindQuery(&q)
	q.Normalize()
	activityType := c.Query("activity_type")
	status := c.Query("status")
	keyword := c.Query("keyword")
	list, total, err := h.svc.List(q.Page, q.PageSize, activityType, status, keyword)
	if err != nil {
		h.wrapError(c, err, "Activity list failed")
		return
	}
	OK(c, pageResponse(list, total, q.Page, q.PageSize))
}

// Calendar 活动日历。
func (h *ActivityHandler) Calendar(c *gin.Context) {
	month := c.Query("month")
	list, err := h.svc.Calendar(month)
	if err != nil {
		h.wrapError(c, err, "Activity calendar failed")
		return
	}
	OK(c, list)
}

// Get 活动详情。
func (h *ActivityHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Activity[id] get: invalid id")
		return
	}
	a, count, err := h.svc.Get(id)
	if err != nil {
		h.wrapError(c, err, "Activity get failed")
		return
	}
	OK(c, gin.H{"activity": a, "registered_count": count})
}

// Create 创建活动。
func (h *ActivityHandler) Create(c *gin.Context) {
	var req dto.ActivityCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Activity create: "+err.Error())
		return
	}
	a, err := h.svc.Create(middleware.GetUserID(c), req.Title, req.Description, req.CoverImage,
		req.ActivityType, req.Location, req.StartTime, req.EndTime, req.SignupDeadline, req.Capacity, req.Status)
	if err != nil {
		h.wrapError(c, err, "Activity create failed")
		return
	}
	OKWithMessage(c, "活动创建成功", a)
}

// Update 更新活动。
func (h *ActivityHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Activity[id] update: invalid id")
		return
	}
	var req dto.ActivityUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Activity[id="+strconv.FormatUint(id, 10)+"] update: "+err.Error())
		return
	}
	fields := map[string]any{"title": req.Title, "description": req.Description, "cover_image": req.CoverImage,
		"activity_type": req.ActivityType, "location": req.Location}
	if req.Capacity != nil {
		fields["capacity"] = *req.Capacity
	}
	a, err := h.svc.Update(id, middleware.GetUserID(c), middleware.GetUserRole(c), fields)
	if err != nil {
		h.wrapError(c, err, "Activity update failed")
		return
	}
	OK(c, a)
}

// Publish 发布活动。
func (h *ActivityHandler) Publish(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Activity[id] publish: invalid id")
		return
	}
	a, err := h.svc.Publish(id, middleware.GetUserID(c), middleware.GetUserRole(c))
	if err != nil {
		h.wrapError(c, err, "Activity publish failed")
		return
	}
	OKWithMessage(c, constants.MsgActivityPublished, a)
}

// End 结束活动。
func (h *ActivityHandler) End(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Activity[id] end: invalid id")
		return
	}
	a, err := h.svc.End(id, middleware.GetUserID(c), middleware.GetUserRole(c))
	if err != nil {
		h.wrapError(c, err, "Activity end failed")
		return
	}
	OKWithMessage(c, constants.MsgActivityEndedAction, a)
}

// Delete 删除活动。
func (h *ActivityHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Activity[id] delete: invalid id")
		return
	}
	if err := h.svc.Delete(id, middleware.GetUserID(c), middleware.GetUserRole(c)); err != nil {
		h.wrapError(c, err, "Activity delete failed")
		return
	}
	OK(c, nil)
}

// Mine 我的活动（组织者）。
func (h *ActivityHandler) Mine(c *gin.Context) {
	var q dto.PageQuery
	_ = c.ShouldBindQuery(&q)
	q.Normalize()
	status := c.Query("status")
	list, total, err := h.svc.ListByOrganizer(middleware.GetUserID(c), q.Page, q.PageSize, status)
	if err != nil {
		h.wrapError(c, err, "Activity mine failed")
		return
	}
	OK(c, pageResponse(list, total, q.Page, q.PageSize))
}

// Stats 活动报名/签到统计。
func (h *ActivityHandler) Stats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Activity[id] stats: invalid id")
		return
	}
	stats, err := h.svc.Stats(id, middleware.GetUserID(c), middleware.GetUserRole(c))
	if err != nil {
		h.wrapError(c, err, "Activity stats failed")
		return
	}
	OK(c, stats)
}

func (h *ActivityHandler) wrapError(c *gin.Context, err error, ctx string) {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		c.Set("audit_detail", appErr.Message)
		h.logger.Warn("activity handler error", "context", ctx, "error", appErr.Error())
		Fail(c, appErrorStatus(appErr.Code), appErr.Code, appErr.Message)
		return
	}
	h.logger.Error("activity handler error", "context", ctx, "error", err.Error())
	Fail(c, http.StatusInternalServerError, constants.CodeInternalError, constants.MsgInternalError)
}
