package adapters

import (
	"log/slog"

	"github.com/ekalinin/github-markdown-toc.go/internal/core/ports"
)

type Logger struct {
	debug bool
	log   ports.Logger
}

func NewLogger(debug bool) *Logger {
	return NewLoggerX(debug, slog.Default())
}

func NewLoggerX(debug bool, logger ports.Logger) *Logger {
	return &Logger{
		debug: debug,
		log:   logger,
	}
}

func (l *Logger) Info(format string, v ...any) {
	if l.debug {
		l.log.Info(format, v...)
	}
}
