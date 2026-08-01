package controller

import (
	"context"
	"io"
	"os"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
)

type useCase interface {
	Do(context.Context, string) (entity.Toc, error)
}

type logger interface {
	Info(string, ...any)
}

type Controller struct {
	cfg          Config
	ucLocalMd    useCase
	ucRemoteMD   useCase
	ucRemoteHTML useCase
	log          logger
}

func New(cfg Config, ucLocalMD useCase, ucRemoteMD useCase, ucRemoteHTML useCase, log logger) *Controller {
	return &Controller{
		cfg:          cfg,
		ucLocalMd:    ucLocalMD,
		ucRemoteMD:   ucRemoteMD,
		ucRemoteHTML: ucRemoteHTML,
		log:          log,
	}
}

func (ctl *Controller) Process(ctx context.Context, stdout io.Writer) error {
	if len(ctl.cfg.Files) > 0 {
		return ctl.ProcessFiles(ctx, stdout, ctl.cfg.Files...)
	}
	return ctl.ProcessSTDIN(ctx, stdout, os.Stdin)
}
