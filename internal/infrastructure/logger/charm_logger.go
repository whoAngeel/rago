package logger

import (
	"os"

	charmlog "github.com/charmbracelet/log"
	"github.com/whoAngeel/rago/internal/core/ports"
)

type CharmLogger struct {
	logger *charmlog.Logger
}

func New(env string) ports.Logger {
	l := charmlog.NewWithOptions(os.Stdout, charmlog.Options{
		ReportTimestamp: true,
		TimeFormat:      "15:04:05",
	})

	if env == "production" {
		l.SetFormatter(charmlog.JSONFormatter)
		l.SetLevel(charmlog.InfoLevel)
	} else {
		l.SetFormatter(charmlog.TextFormatter)
		l.SetLevel(charmlog.DebugLevel)
	}

	return &CharmLogger{
		logger: l,
	}
}

func (l *CharmLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *CharmLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *CharmLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *CharmLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *CharmLogger) Fatal(msg string, args ...any) {
	l.logger.Fatal(msg, args...)
}

func (l *CharmLogger) With(args ...any) ports.Logger {
	return &CharmLogger{
		logger: l.logger.With(args...),
	}
}
