package adapters

import (
	"fmt"
	"io"
)

// Notifier writes status messages for the user. They go to a stream separate from
// the TOC itself, because the worker pool emits them in completion order.
type Notifier struct {
	w io.Writer
}

func NewNotifier(w io.Writer) *Notifier {
	return &Notifier{w: w}
}

func (n *Notifier) Notify(format string, args ...any) {
	if n.w == nil {
		return
	}
	_, _ = fmt.Fprintf(n.w, format+"\n", args...)
}
