package filter

import (
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// AccessLogMiddleware logs request summary after response is written.
func AccessLogMiddleware(log *slog.Logger) gin.HandlerFunc {
	if log == nil {
		log = slog.Default()
	}
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		ctx := FromGin(c)
		rid := ctx.RequestID
		if rid == "" {
			rid = c.GetHeader("X-Request-Id")
		}
		log.Info("access",
			slog.String("request_id", rid),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Int64("latency_ms", time.Since(start).Milliseconds()),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}

// NewAccessLogger returns a slog logger writing to stdout and optional file.
func NewAccessLogger(logPath string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	h := slog.NewTextHandler(os.Stdout, opts)
	if logPath == "" {
		return slog.New(h)
	}
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return slog.New(h)
	}
	mw := io.MultiWriter(os.Stdout, f)
	return slog.New(slog.NewJSONHandler(mw, opts))
}
