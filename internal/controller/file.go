package controller

import (
	"errors"
	"io"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
)

func (ctl *Controller) getUseCase(file string) useCase {
	switch t := entity.GetType(file); t {
	case entity.TypeLocalMD:
		ctl.log.Info("Controller.ProcessFiles: detect use-case", "use-case", entity.TypeLocalMD)
		return ctl.ucLocalMd
	case entity.TypeRemoteMD:
		ctl.log.Info("Controller.ProcessFiles: detect use-case", "use-case", entity.TypeRemoteMD)
		return ctl.ucRemoteMD
	case entity.TypeRemoteHTML:
		ctl.log.Info("Controller.ProcessFiles: detect use-case", "use-case", entity.TypeRemoteHTML)
		return ctl.ucRemoteHTML
	}
	ctl.log.Info("Controller.ProcessFiles: use-case is null")
	return nil
}

func (ctl *Controller) ProcessFiles(stdout io.Writer, files ...string) error {
	ctl.log.Info("Controller.ProcessFiles: start", "files", files)
	cnt := len(files)

	ch := make(chan *entity.Toc, cnt)
	for _, file := range files {
		ctl.log.Info("Controller.ProcessFiles: processing", "file", file)
		uc := ctl.getUseCase(file)
		if uc == nil {
			return errors.New("useCase is null")
		}

		if ctl.cfg.Serial {
			ch <- uc.Do(file)
		} else {
			go func(ucc useCase, path string) {
				ch <- ucc.Do(path)
			}(uc, file)
		}
	}

	for i := 0; i < cnt; i++ {
		toc := <-ch
		// #14, check if there's really TOC?
		if toc != nil {
			if err := toc.Print(stdout); err != nil {
				return err
			}
		}
	}
	return nil
}
