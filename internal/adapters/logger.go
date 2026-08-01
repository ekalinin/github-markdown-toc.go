package adapters

import (
	"log/slog"
)

type logger interface {
	Info(string, ...any)
}

type Logger struct {
	debug bool
	log   logger
}

func NewLogger(debug bool) *Logger {
	return NewLoggerX(debug, slog.Default())
}

func NewLoggerX(debug bool, logger logger) *Logger {
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
