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

	type result struct {
		file string
		toc  *entity.Toc
	}
	ch := make(chan result, cnt)

	for _, file := range files {
		ctl.log.Info("Controller.ProcessFiles: processing", "file", file)
		uc := ctl.getUseCase(file)
		if uc == nil {
			return errors.New("useCase is null")
		}

		if ctl.cfg.Serial {
			ch <- result{file: file, toc: uc.Do(file)}
		} else {
			go func(ucc useCase, path string) {
				ch <- result{file: path, toc: ucc.Do(path)}
			}(uc, file)
		}
	}

	for i := 0; i < cnt; i++ {
		res := <-ch
		// #14, check if there's really TOC?
		if res.toc != nil {
			if ctl.cfg.Insert {
				if entity.GetType(res.file) != entity.TypeLocalMD {
					ctl.log.Info("Skipping insert for non-local file, printing to stdout instead", "file", res.file)
					if err := res.toc.Print(stdout); err != nil {
						return err
					}
					continue
				}

				if err := ctl.insertTocToFile(res.file, res.toc); err != nil {
					return err
				}
				if err := res.toc.Print(stdout); err != nil {
					return err
				}
			} else {
				if err := res.toc.Print(stdout); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (ctl *Controller) insertTocToFile(filepath string, toc *entity.Toc) error {
	if !ctl.cfg.NoBackup {
		backupPath, err := ctl.fileBackupper.CreateBackup(filepath)
		if err != nil {
			ctl.log.Info("Failed to create backup", "error", err)
			return err
		}
		ctl.log.Info("Created backup", "path", backupPath)
	}

	tocStr := toc.String()
	if err := ctl.tocInserter.InsertToc(filepath, tocStr); err != nil {
		ctl.log.Info("Failed to insert TOC", "error", err)
		return err
	}

	ctl.log.Info("TOC inserted", "file", filepath)
	return nil
}
