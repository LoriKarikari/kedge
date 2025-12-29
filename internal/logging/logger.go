package logging

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

type Config struct {
	Level  string
	Format string
}

func New(cfg Config) *slog.Logger {
	level := slog.LevelInfo
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	if cfg.Format == "json" {
		return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	}
	return slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: level}))
}
