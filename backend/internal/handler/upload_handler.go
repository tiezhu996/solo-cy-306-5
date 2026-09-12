package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"gbevent/internal/config"
	"gbevent/internal/constants"
	"gbevent/internal/util"

	"github.com/gin-gonic/gin"
)

// UploadHandler 文件上传处理器。
type UploadHandler struct {
	cfg    *config.Config
	logger *slog.Logger
}

// NewUploadHandler 构造上传处理器。
func NewUploadHandler(cfg *config.Config, logger *slog.Logger) *UploadHandler {
	return &UploadHandler{cfg: cfg, logger: logger}
}

// UploadImage 上传图片。
func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Upload image: missing file")
		return
	}
	url, err := util.SaveUploadedImage(h.cfg.UploadDir, h.cfg.UploadMaxMB, file)
	if err != nil {
		h.logger.Error(constants.LogUploadImageFailed, "error", err.Error())
		var appErr *util.AppError
		if errors.As(err, &appErr) {
			Fail(c, appErrorStatus(appErr.Code), appErr.Code, appErr.Message)
			return
		}
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, "Upload image failed: "+err.Error())
		return
	}
	h.logger.Info(constants.LogUploadImageSuccess, "url", url)
	OK(c, gin.H{"url": url})
}
