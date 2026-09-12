package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"gbevent/internal/constants"
	"gbevent/internal/util"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 统一错误响应格式。
func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err
		var appErr *util.AppError
		if errors.As(err, &appErr) {
			status := http.StatusBadRequest
			switch appErr.Code {
			case constants.CodeUnauthorized, constants.CodeInvalidCredentials:
				status = http.StatusUnauthorized
			case constants.CodeForbidden:
				status = http.StatusForbidden
			case constants.CodeNotFound, constants.CodeUserNotFound, constants.CodeInvalidVoucher:
				status = http.StatusNotFound
			case constants.CodeConflict, constants.CodeActivityFull, constants.CodeDuplicateSignup,
				constants.CodeAlreadyCheckedIn, constants.CodeReviewConflict, constants.CodeCancelConflict:
				status = http.StatusConflict
			case constants.CodeTooManyRequests:
				status = http.StatusTooManyRequests
			case constants.CodeValidationFailed:
				status = http.StatusUnprocessableEntity
			case constants.CodeUploadTooLarge:
				status = http.StatusRequestEntityTooLarge
			case constants.CodeUnsupportedType:
				status = http.StatusUnsupportedMediaType
			default:
				status = http.StatusInternalServerError
			}
			c.JSON(status, gin.H{"code": appErr.Code, "message": appErr.Message, "data": nil})
			return
		}
		logger.Error("unhandled error", "error", err.Error(), "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"code": constants.CodeInternalError, "message": constants.MsgInternalError, "data": nil})
	}
}
