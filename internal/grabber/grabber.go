package grabber

import (
	ghtoc "github.com/ekalinin/github-markdown-toc.go"
	"github.com/ekalinin/github-markdown-toc.go/internal/debugger"
)

type Grabber interface {
	Grab(html string) (ghtoc.GHToc, error)
}

type DefaultGrabber struct {
	Path string

	// toc grabber
	AbsPaths   bool
	StartDepth int
	Depth      int
	Escape     bool
	Indent     int

	// internals
	debugger.Debugger
}

// TODO: add option functions for settings
