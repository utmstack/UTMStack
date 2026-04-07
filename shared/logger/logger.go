// Package logger provides a simple logging wrapper.
package logger

import (
	"github.com/threatwinds/logger"
)

// Log level constants
const (
	LevelDebug    = 100
	LevelInfo     = 200
	LevelNotice   = 300
	LevelWarning  = 400
	LevelError    = 500
	LevelCritical = 502
	LevelAlert    = 509
)

// Logger is the shared logger instance.
var Logger *logger.Logger

// Init initializes the logger with the specified log file path and level.
func Init(logFile string, level int) {
	Logger = logger.NewLogger(&logger.Config{
		Format:  "text",
		Level:   level,
		Output:  logFile,
		Retries: 3,
		Wait:    5,
	})
}

// Info logs an informational message.
func Info(format string, args ...any) {
	if Logger != nil {
		Logger.Info(format, args...)
	}
}

// Error logs an error message.
func Error(format string, args ...any) {
	if Logger != nil {
		Logger.ErrorF(format, args...)
	}
}

// Fatal logs a fatal message and exits.
func Fatal(format string, args ...any) {
	if Logger != nil {
		Logger.Fatal(format, args...)
	}
}

// Debug logs a debug message with the specified level.
func Debug(level int, format string, args ...any) {
	if Logger != nil {
		Logger.LogF(level, format, args...)
	}
}
