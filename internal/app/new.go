package app

import (
	"github.com/ekalinin/github-markdown-toc.go/internal/adapters"
	"github.com/ekalinin/github-markdown-toc.go/internal/controller"
	"github.com/ekalinin/github-markdown-toc.go/internal/core/usecase"
)

type Controller interface {
	Process() error
}

type App struct {
	cfg Config
	ctl Controller
}

func New(cfg Config) *App {
	log := adapters.NewLogger(cfg.Debug)

	log.Info("App.New: init configs ...")
	ctlCfg := cfg.ToControllerConfig()
	ucCfg := ctlCfg.ToUseCaseConfig()

	log.Info("App.New: init adapters ...")
	checker := adapters.NewFileCheck(log)
	writer := adapters.NewFileWriter(log)
	converter := adapters.NewHTMLConverter(cfg.GHToken, cfg.GHUrl, log)
	grabber := adapters.NewReGrabber("", cfg.ToGrabberConfig(), cfg.GHVersion)

	log.Info("App.New: init usecases ...")
	ucLocalMD, ucRemoteMD, ucRemoteHTML := usecase.New(
		ucCfg, checker, writer, converter, grabber, log,
	)
	// TODO: set JSONGrabber for remoreTHML

	log.Info("App.New: init controller ...")
	ctl := controller.New(ctlCfg, ucLocalMD, ucRemoteMD, ucRemoteHTML, log)

	log.Info("done.")
	return &App{
		ctl: ctl,
		cfg: cfg,
	}
}
