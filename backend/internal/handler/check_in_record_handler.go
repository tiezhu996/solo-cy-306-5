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

// CheckInRecordHandler 签到 HTTP 处理器。
type CheckInRecordHandler struct {
	svc    *service.CheckInRecordService
	logger *slog.Logger
}

// NewCheckInRecordHandler 构造签到处理器。
func NewCheckInRecordHandler(svc *service.CheckInRecordService, logger *slog.Logger) *CheckInRecordHandler {
	return &CheckInRecordHandler{svc: svc, logger: logger}
}

// CheckIn 凭证/扫码签到。
func (h *CheckInRecordHandler) CheckIn(c *gin.Context) {
	var req dto.CheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "CheckInRecord checkin: "+err.Error())
		return
	}
	if req.Voucher == "" && req.QRContent == "" {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "CheckInRecord checkin: voucher or qr_content required")
		return
	}
	activityID, _ := strconv.ParseUint(c.Query("activity_id"), 10, 64)
	if activityID == 0 {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "CheckInRecord checkin: invalid activity_id")
		return
	}
	var rec any
	var err error
	if req.QRContent != "" {
		rec, err = h.svc.CheckInByScan(activityID, middleware.GetUserID(c), req.QRContent)
	} else {
		rec, err = h.svc.CheckInByVoucher(activityID, middleware.GetUserID(c), req.Voucher)
	}
	if err != nil {
		h.wrapError(c, err, "CheckInRecord checkin failed")
		return
	}
	OKWithMessage(c, constants.MsgCheckinSuccess, rec)
}

// Records 某活动签到记录。
func (h *CheckInRecordHandler) Records(c *gin.Context) {
	activityID, err := strconv.ParseUint(c.Query("activity_id"), 10, 64)
	if err != nil || activityID == 0 {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "CheckInRecord records: invalid activity_id")
		return
	}
	list, err := h.svc.ListByActivity(activityID)
	if err != nil {
		h.wrapError(c, err, "CheckInRecord records failed")
		return
	}
	OK(c, list)
}

func (h *CheckInRecordHandler) wrapError(c *gin.Context, err error, ctx string) {
	var appErr *util.AppError
	if errors.As(err, &appErr) {
		h.logger.Warn("checkin handler error", "context", ctx, "error", appErr.Error())
		Fail(c, appErrorStatus(appErr.Code), appErr.Code, appErr.Message)
		return
	}
	h.logger.Error("checkin handler error", "context", ctx, "error", err.Error())
	Fail(c, http.StatusInternalServerError, constants.CodeInternalError, constants.MsgInternalError)
}
