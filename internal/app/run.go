package app

import (
	"io"

	"github.com/ekalinin/github-markdown-toc.go/internal/utils"
)

func (a *App) Run(stdout io.Writer) error {

	// do not show for stdin case (Files is empty)
	if !a.cfg.HideHeader && len(a.cfg.Files) == 1 {
		utils.ShowHeader(stdout)
	}

	if err := a.ctl.Process(stdout); err != nil {
		return err
	}

	if !a.cfg.HideFooter {
		utils.ShowFooter(stdout)
	}

	return nil
}
