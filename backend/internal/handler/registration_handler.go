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

// RegistrationHandler 报名 HTTP 处理器。
type RegistrationHandler struct {
	svc    *service.RegistrationService
	logger *slog.Logger
}

// NewRegistrationHandler 构造报名处理器。
func NewRegistrationHandler(svc *service.RegistrationService, logger *slog.Logger) *RegistrationHandler {
	return &RegistrationHandler{svc: svc, logger: logger}
}

// Create 在线报名。
func (h *RegistrationHandler) Create(c *gin.Context) {
	var req dto.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Registration create: "+err.Error())
		return
	}
	reg, err := h.svc.Create(req.ActivityID, middleware.GetUserID(c), req.Name, req.Phone, req.Remark)
	if err != nil {
		h.wrapError(c, err, "Registration create failed")
		return
	}
	OKWithMessage(c, constants.MsgSignupSuccess, reg)
}

// List 报名列表（组织者）。
func (h *RegistrationHandler) List(c *gin.Context) {
	var q dto.PageQuery
	_ = c.ShouldBindQuery(&q)
	q.Normalize()
	activityID, _ := strconv.ParseUint(c.Query("activity_id"), 10, 64)
	status := c.Query("status")
	reviewStatus := c.Query("review_status")
	list, total, err := h.svc.List(q.Page, q.PageSize, activityID, status, reviewStatus)
	if err != nil {
		h.wrapError(c, err, "Registration list failed")
		return
	}
	OK(c, pageResponse(list, total, q.Page, q.PageSize))
}

// Mine 我的报名。
func (h *RegistrationHandler) Mine(c *gin.Context) {
	var q dto.PageQuery
	_ = c.ShouldBindQuery(&q)
	q.Normalize()
	list, total, err := h.svc.ListMine(middleware.GetUserID(c), q.Page, q.PageSize)
	if err != nil {
		h.wrapError(c, err, "Registration mine failed")
		return
	}
	OK(c, pageResponse(list, total, q.Page, q.PageSize))
}

// Cancel 取消报名。
func (h *RegistrationHandler) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Registration[id] cancel: invalid id")
		return
	}
	reg, err := h.svc.Cancel(id, middleware.GetUserID(c), middleware.GetUserRole(c))
	if err != nil {
		h.wrapError(c, err, "Registration cancel failed")
		return
	}
	OKWithMessage(c, constants.MsgCancelSuccess, reg)
}

// Review 审核报名。
func (h *RegistrationHandler) Review(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Registration[id] review: invalid id")
		return
	}
	var req dto.ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Registration[id="+strconv.FormatUint(id, 10)+"] review: "+err.Error())
		return
	}
	reg, err := h.svc.Review(id, middleware.GetUserID(c), middleware.GetUserRole(c), req.ReviewStatus)
	if err != nil {
		h.wrapError(c, err, "Registration review failed")
		return
	}
	OK(c, reg)
}

// OfflineCreate 线下补录。
func (h *RegistrationHandler) OfflineCreate(c *gin.Context) {
	var req dto.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Registration offline create: "+err.Error())
		return
	}
	reg, err := h.svc.OfflineCreate(req.ActivityID, middleware.GetUserID(c), req.Name, req.Phone, req.Remark)
	if err != nil {
		h.wrapError(c, err, "Registration offline create failed")
		return
	}
	OKWithMessage(c, "线下补录成功", reg)
}

// Export 导出报名名单 CSV。
func (h *RegistrationHandler) Export(c *gin.Context) {
	activityID, err := strconv.ParseUint(c.Query("activity_id"), 10, 64)
	if err != nil || activityID == 0 {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Registration export: invalid activity_id")
		return
	}
	filename, content, err := h.svc.ExportCSV(activityID, middleware.GetUserID(c), middleware.GetUserRole(c))
	if err != nil {
		h.wrapError(c, err, "Registration export failed")
		return
	}
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.String(http.StatusOK, content)
}

func (h *RegistrationHandler) wrapError(c *gin.Context, err error, ctx string) {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		c.Set("audit_detail", appErr.Message)
		h.logger.Warn("registration handler error", "context", ctx, "error", appErr.Error())
		Fail(c, appErrorStatus(appErr.Code), appErr.Code, appErr.Message)
		return
	}
	h.logger.Error("registration handler error", "context", ctx, "error", err.Error())
	Fail(c, http.StatusInternalServerError, constants.CodeInternalError, constants.MsgInternalError)
}
