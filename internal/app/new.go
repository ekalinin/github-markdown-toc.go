package app

import (
	"context"
	"fmt"
	"io"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/adapters"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/controller"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase"
)

type Controller interface {
	Process(ctx context.Context, stdout io.Writer) error
}

type App struct {
	cfg Config
	ctl Controller
}

func New(cfg Config) (*App, error) {
	log := adapters.NewLogger(cfg.Debug)

	log.Info(
		"App.New: init configs ...",
		"file-count", len(cfg.Files),
		"serial", cfg.Serial,
		"hide-header", cfg.HideHeader,
		"hide-footer", cfg.HideFooter,
		"start-depth", cfg.StartDepth,
		"depth", cfg.Depth,
		"no-escape", cfg.NoEscape,
		"indent", cfg.Indent,
		"github-version", cfg.GHVersion,
		"token-configured", cfg.GHToken != "",
	)
	ctlCfg := cfg.ToControllerConfig()
	ucCfg := ctlCfg.ToUseCaseConfig()

	log.Info("App.New: init adapters ...")
	checker := adapters.NewFileCheck(log)
	writer := adapters.NewFileWriter(log)
	converter := adapters.NewHTMLConverter(cfg.GHToken, cfg.GHUrl, log)
	grabberRe, err := adapters.NewReGrabber("", cfg.ToGrabberConfig(), cfg.GHVersion)
	if err != nil {
		return nil, fmt.Errorf("initialize regexp grabber: %w", err)
	}
	grabberJson := adapters.NewJsonGrabber(cfg.ToGrabberConfig())
	getter := adapters.NewRemoteGetter(true)
	temper := adapters.NewFileTemper()

	log.Info("App.New: init usecases ...")
	ucLocalMD, ucRemoteMD, ucRemoteHTML := usecase.New(
		ucCfg, checker, writer, converter, grabberRe, grabberJson,
		getter, temper, log,
	)

	log.Info("App.New: init controller ...")
	ctl := controller.New(ctlCfg, ucLocalMD, ucRemoteMD, ucRemoteHTML, log)

	log.Info("App.New: done.")
	return &App{
		ctl: ctl,
		cfg: cfg,
	}, nil
}
