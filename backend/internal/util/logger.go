package util

import (
	"log/slog"
	"os"
)

// NewLogger 创建 slog 结构化日志器。
func NewLogger(level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
