package controller

import (
	"io"
	"os"

	"github.com/ekalinin/github-markdown-toc.go/internal/core/entity"
	"github.com/ekalinin/github-markdown-toc.go/internal/core/ports"
)

type useCase interface {
	Do(string) *entity.Toc
}

type Controller struct {
	cfg          Config
	ucLocalMd    useCase
	ucRemoteMD   useCase
	ucRemoteHTML useCase
	log          ports.Logger
}

func New(cfg Config, ucLocalMD useCase, ucRemoteMD useCase, ucRemoteHTML useCase, log ports.Logger) *Controller {
	return &Controller{
		cfg:          cfg,
		ucLocalMd:    ucLocalMD,
		ucRemoteMD:   ucRemoteMD,
		ucRemoteHTML: ucRemoteHTML,
		log:          log,
	}
}

func (ctl *Controller) Process(stdout io.Writer) error {
	if len(ctl.cfg.Files) > 0 {
		return ctl.ProcessFiles(stdout, ctl.cfg.Files...)
	}
	return ctl.ProcessSTDIN(stdout, os.Stdin)
}
