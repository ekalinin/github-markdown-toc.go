package app

import (
	"context"
	"fmt"
	"io"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/adapters"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/controller"
	coretoc "github.com/ekalinin/github-markdown-toc.go/v2/internal/core/toc"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase/localmd"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase/remotehtml"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase/remotemd"
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
		"hide-header", cfg.Presentation.HideHeader,
		"hide-footer", cfg.Presentation.HideFooter,
		"start-depth", cfg.TOC.StartDepth,
		"depth", cfg.TOC.Depth,
		"no-escape", !cfg.TOC.Escape,
		"indent", cfg.TOC.Indent,
		"github-version", cfg.GitHub.GHVersion,
		"token-configured", cfg.GitHub.GHToken != "",
	)
	ctlCfg := controller.Config{Files: cfg.Files, Serial: cfg.Serial}

	log.Info("App.New: init adapters ...")
	if cfg.GitHub.GHToken == "" {
		token, err := adapters.NewTokenResolver().Resolve()
		if err != nil {
			return nil, fmt.Errorf("read token file: %w", err)
		}
		cfg.GitHub.GHToken = token
	}
	httpClient := adapters.NewHTTPClient()
	checker := adapters.NewFileCheck(log)
	writer := adapters.NewFileWriter()
	converter := adapters.NewHTMLConverterWithClient(cfg.GitHub.GHToken, cfg.GitHub.GHUrl, httpClient, log)
	regexpExtractor, err := adapters.NewRegexpExtractor(cfg.GitHub.GHVersion)
	if err != nil {
		return nil, fmt.Errorf("initialize regexp grabber: %w", err)
	}
	jsonExtractor := adapters.NewJSONExtractor()
	rendererCfg := cfg.TOC
	// bash gh-md-toc drops the path prefix only when a single document is requested.
	rendererCfg.AbsolutePaths = len(cfg.Files) > 1
	renderer := coretoc.NewRenderer(rendererCfg)
	grabberRe := coretoc.NewGenerator(regexpExtractor, renderer)
	grabberJSON := coretoc.NewGenerator(jsonExtractor, renderer)
	getter := adapters.NewRemoteGetterWithClient(true, httpClient)
	temper := adapters.NewFileTemper()

	log.Info("App.New: init usecases ...")
	ucLocalMD := localmd.New(cfg.Debug, checker, writer, converter, grabberRe, log)
	ucRemoteMD := remotemd.New(getter, ucLocalMD, temper, log)
	ucRemoteHTML := remotehtml.New(cfg.Debug, getter, temper, grabberJSON, log)

	log.Info("App.New: init controller ...")
	ctl := controller.New(ctlCfg, ucLocalMD, ucRemoteMD, ucRemoteHTML, log)

	log.Info("App.New: done.")
	return &App{
		ctl: ctl,
		cfg: cfg,
	}, nil
}
