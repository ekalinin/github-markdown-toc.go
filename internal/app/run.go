package app

import (
	"context"
	"fmt"
	"io"
)

func (a *App) Run(ctx context.Context, stdout io.Writer) error {

	// do not show for stdin case (Files is empty)
	if !a.cfg.Presentation.HideHeader && len(a.cfg.Files) == 1 {
		ShowHeader(stdout)
	}

	if err := a.ctl.Process(ctx, stdout); err != nil {
		return err
	}

	if !a.cfg.Presentation.HideFooter {
		ShowFooter(stdout)
	}

	return nil
}

// ShowHeader writes the heading shown before a generated TOC.
func ShowHeader(w io.Writer) {
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Table of Contents")
	_, _ = fmt.Fprintln(w, "=================")
	_, _ = fmt.Fprintln(w)
}

// ShowFooter writes the attribution shown after a generated TOC.
func ShowFooter(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)")
}
