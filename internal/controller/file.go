package controller

import (
	"context"
	"errors"
	"fmt"
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

type processResult struct {
	path string
	toc  entity.Toc
	err  error
}

func (ctl *Controller) ProcessFiles(ctx context.Context, stdout io.Writer, files ...string) error {
	ctl.log.Info("Controller.ProcessFiles: start", "files", files)
	cnt := len(files)

	ch := make(chan processResult, cnt)
	for _, file := range files {
		ctl.log.Info("Controller.ProcessFiles: processing", "file", file)
		uc := ctl.getUseCase(file)
		if uc == nil {
			ch <- processResult{
				path: file,
				err:  fmt.Errorf("process %q: select use case: use case is nil", file),
			}
			continue
		}

		process := func(ctx context.Context, uc useCase, path string) processResult {
			if err := ctx.Err(); err != nil {
				return processResult{
					path: path,
					err:  fmt.Errorf("process %q: %w", path, err),
				}
			}

			toc, err := uc.Do(ctx, path)
			if err != nil {
				err = fmt.Errorf("process %q: %w", path, err)
			}
			return processResult{path: path, toc: toc, err: err}
		}

		if ctl.cfg.Serial {
			ch <- process(ctx, uc, file)
		} else {
			go func(ctx context.Context, uc useCase, path string) {
				ch <- process(ctx, uc, path)
			}(ctx, uc, file)
		}
	}

	var errs []error
	for i := 0; i < cnt; i++ {
		result := <-ch
		if result.err != nil {
			errs = append(errs, result.err)
			continue
		}
		if err := result.toc.Print(stdout); err != nil {
			errs = append(errs, fmt.Errorf("print TOC for %q: %w", result.path, err))
		}
	}
	return errors.Join(errs...)
}
