package file

import (
	"fmt"

	"github.com/ekalinin/github-markdown-toc.go/internal/debugger"
)

type file struct {
	Path    string
	GhToken string
	GhUrl   string

	// toc grabber
	AbsPaths   bool
	StartDepth int
	Depth      int
	Escape     bool
	Indent     int

	// internals
	debugger.Debugger
}

func (f file) ShowTocHeader() {
	fmt.Println()
	fmt.Println("Table of Contents")
	fmt.Println("=================")
	fmt.Println()
}
