package app

import (
	"io"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/adapters"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/controller"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase"
)

type Controller interface {
	Process(stdout io.Writer) error
}

type App struct {
	cfg Config
	ctl Controller
}

func New(cfg Config) *App {
	log := adapters.NewLogger(cfg.Debug)

	log.Info("App.New: init configs ...", "app cfg", cfg)
	ctlCfg := cfg.ToControllerConfig()
	ucCfg := ctlCfg.ToUseCaseConfig()

	log.Info("App.New: init adapters ...")
	checker := adapters.NewFileCheck(log)
	writer := adapters.NewFileWriter(log)
	converter := adapters.NewHTMLConverter(cfg.GHToken, cfg.GHUrl, log)
	grabberRe := adapters.NewReGrabber("", cfg.ToGrabberConfig(), cfg.GHVersion)
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
	}
}
