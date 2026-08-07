package utils

import (
"log/slog"
"os"
)

// Logger is the application-wide structured logger.
var Logger *slog.Logger

func init() {
	Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// InitLogger initializes the global logger. Call once at app startup.
func InitLogger() {
	Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// Info logs at info level with contextual fields.
func Info(msg string, args ...any) {
	Logger.Info(msg, args...)
}

// Error logs at error level with contextual fields.
func Error(msg string, args ...any) {
	Logger.Error(msg, args...)
}

// Warn logs at warn level with contextual fields.
func Warn(msg string, args ...any) {
	Logger.Warn(msg, args...)
}

// Debug logs at debug level with contextual fields.
func Debug(msg string, args ...any) {
	Logger.Debug(msg, args...)
}
