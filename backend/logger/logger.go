package logger

import (
	"log/slog"
	"os"
)

var (
	log *slog.Logger
)

func Init(level string) {
	var lvl slog.Level
	switch level {
	case "DEBUG":
		lvl = slog.LevelDebug
	case "INFO":
		lvl = slog.LevelInfo
	case "WARN":
		lvl = slog.LevelWarn
	case "ERROR":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     lvl,
		AddSource: true,
	}))
	slog.SetDefault(log)
}

func Get() *slog.Logger {
	if log == nil {
		Init("INFO")
	}
	return log
}

func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

func Fatal(msg string, args ...any) {
	Get().Error(msg, args...)
	os.Exit(1)
}
