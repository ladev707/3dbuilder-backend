package logger

import (
	"os"
	"log/slog"
)

var Logger *slog.Logger

func InitLogger() {
	var level slog.Level
	switch os.Getenv("LOG_LEVEL") {
	case "INFO":
		level = slog.LevelInfo
	case "DEBUG":
		level = slog.LevelDebug
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	}))
	slog.SetDefault(Logger)
}

func Info(msg string, keyvals ...any) {
	Logger.Info(msg, keyvals...)
}

func Error(msg string, keyvals ...any) {
	Logger.Error(msg, keyvals...)
}

func Debug(msg string, keyvals ...any) {
	Logger.Debug(msg, keyvals...)
}
