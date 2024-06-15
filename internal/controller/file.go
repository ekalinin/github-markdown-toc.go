package controller

import (
	"errors"
	"os"

	"github.com/ekalinin/github-markdown-toc.go/internal/core/entity"
)

func (ctl *Controller) getUseCase(file string) useCase {
	switch t := entity.GetType(file); t {
	case entity.TypeLocalMD:
		return ctl.ucLocalMd
	case entity.TypeRemoteMD:
		return ctl.ucRemoteMD
	case entity.TypeRemoteHTML:
		return ctl.ucRemoteHTML
	}
	return nil
}

func (ctl *Controller) ProcessFiles(files ...string) error {
	ctl.log.Info("Controller.ProcessFiles: start", "files", files)
	cnt := len(files)

	ch := make(chan *entity.Toc, cnt)
	for _, file := range files {
		ctl.log.Info("Controller.ProcessFiles: processing", "file", file)
		useCase := ctl.getUseCase(file)
		if useCase == nil {
			return errors.New("useCase is null")
		}

		if ctl.cfg.Serial {
			ch <- useCase.Do(file)
		} else {
			go func(path string) { ch <- useCase.Do(path) }(file)
		}
	}

	for i := 0; i < cnt; i++ {
		toc := <-ch
		// #14, check if there's really TOC?
		if toc != nil {
			return toc.Print(os.Stdout)
		}
	}
	return nil
}
