package controller

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

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
	index int
	path  string
	toc   entity.Toc
	err   error
}

type processJob struct {
	index int
	path  string
}

const maxParallelFiles = 8

func (ctl *Controller) ProcessFiles(ctx context.Context, stdout io.Writer, files ...string) error {
	ctl.log.Info("Controller.ProcessFiles: start", "files", files)
	jobs := make(chan processJob, len(files))
	resultCh := make(chan processResult, len(files))
	for index, path := range files {
		jobs <- processJob{index: index, path: path}
	}
	close(jobs)

	workerCount := min(len(files), maxParallelFiles)
	if ctl.cfg.Serial {
		workerCount = min(len(files), 1)
	}
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer workers.Done()
			ctl.processFilesWorker(ctx, jobs, resultCh)
		}()
	}

	results := make([]processResult, len(files))
	for range files {
		result := <-resultCh
		results[result.index] = result
	}
	workers.Wait()

	var errs []error
	for _, result := range results {
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

func (ctl *Controller) processFilesWorker(
	ctx context.Context,
	jobs <-chan processJob,
	results chan<- processResult,
) {
	for job := range jobs {
		results <- ctl.processFile(ctx, job)
	}
}

func (ctl *Controller) processFile(ctx context.Context, job processJob) processResult {
	result := processResult{index: job.index, path: job.path}
	if err := ctx.Err(); err != nil {
		result.err = fmt.Errorf("process %q: %w", job.path, err)
		return result
	}

	ctl.log.Info("Controller.ProcessFiles: processing", "file", job.path)
	uc := ctl.getUseCase(job.path)
	if uc == nil {
		result.err = fmt.Errorf("process %q: select use case: use case is nil", job.path)
		return result
	}

	result.toc, result.err = uc.Do(ctx, job.path)
	if result.err != nil {
		result.err = fmt.Errorf("process %q: %w", job.path, result.err)
	}
	return result
}
