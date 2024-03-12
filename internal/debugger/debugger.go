package debugger

import (
	"log"
	"os"
)

type Debugger struct {
	Debug  bool
	logger *log.Logger
}

func New(debug bool, prefix string) Debugger {
	return Debugger{
		Debug:  debug,
		logger: log.New(os.Stderr, prefix, log.LstdFlags),
	}
}

// SetPrefix sets the output prefix for the logger.
func (l *Debugger) SetPrefix(prefix string) {
	l.logger.SetPrefix(prefix)
}

func (d Debugger) Log(format string, v ...any) {
	if d.Debug {
		d.logger.Printf(format+"\n", v...)
	}
}
