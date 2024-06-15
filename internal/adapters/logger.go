package adapters

import "log/slog"

type Logger struct {
	debug bool
	log   *slog.Logger
}

func NewLogger(debug bool) *Logger {
	return &Logger{
		debug: debug,
		log:   slog.Default(),
	}
}

func (l *Logger) Info(format string, v ...any) {
	if l.debug {
		l.log.Info(format, v...)
	}
}
