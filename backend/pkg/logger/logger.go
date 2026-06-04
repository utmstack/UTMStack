package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type Config struct {
	MinLevel    string
	ServiceName string
}

var global *slog.Logger

func Init(cfg *Config) {
	level := slog.LevelInfo
	switch cfg.MinLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	global = slog.New(h).With("service", cfg.ServiceName)
}

func mustGlobal() *slog.Logger {
	if global == nil {
		global = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return global
}

func Debug(msg string)    { mustGlobal().Debug(msg) }
func Info(msg string)     { mustGlobal().Info(msg) }
func Warn(msg string)     { mustGlobal().Warn(msg) }
func Error(msg string)    { mustGlobal().Error(msg) }
func Critical(msg string) { mustGlobal().Error(msg, "level", "CRITICAL") }

func GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		mustGlobal().Info("http",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}
