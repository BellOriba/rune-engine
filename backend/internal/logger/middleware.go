package logger

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Middleware(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		reqID := uuid.New().String()

		c.Writer.Header().Set("X-Request-ID", reqID)

		reqLogger := log.With(
			slog.String("request_id", reqID),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("ip", c.ClientIP()),
		)

		c.Set("logger", reqLogger)

		c.Next()

		status := c.Writer.Status()
		duration := time.Since(start)

		reqLogger.Info("request completed",
			slog.Int("status", status),
			slog.Duration("duration", duration),
		)
	}
}

func FromContext(c *gin.Context) *slog.Logger {
	if l, exists := c.Get("logger"); exists {
		if logger, ok := l.(*slog.Logger); ok {
			return logger
		}
	}

	return slog.Default()
}
