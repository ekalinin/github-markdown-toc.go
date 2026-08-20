package app

import (
	"context"
	"fmt"
	"io"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/adapters"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/controller"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
	coretoc "github.com/ekalinin/github-markdown-toc.go/v2/internal/core/toc"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase/insertmd"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase/localmd"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase/remotehtml"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase/remotemd"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase/skipheader"
)

type Controller interface {
	Process(ctx context.Context, stdout io.Writer) error
}

type useCase interface {
	Do(ctx context.Context, file string) (entity.Toc, error)
}

// markdownProcessor is the local-document pipeline: the plain use case, or one
// wrapped by skipheader. Both forms carry the display path through DoAs.
type markdownProcessor interface {
	Do(ctx context.Context, file string) (entity.Toc, error)
	DoAs(ctx context.Context, file, displayPath string) (entity.Toc, error)
}

type notifier interface {
	Notify(format string, args ...any)
}

type App struct {
	cfg    Config
	ctl    Controller
	notify notifier
}

func New(cfg Config, stderr io.Writer) (*App, error) {
	log := adapters.NewLogger(cfg.Debug)
	notify := adapters.NewNotifier(stderr)

	// Resolve the token.txt fallback before logging, so "token-configured" reflects
	// the token that will actually be used, not just what --token/GH_TOC_TOKEN set.
	if cfg.GitHub.GHToken == "" {
		token, err := adapters.NewTokenResolver().Resolve()
		if err != nil {
			return nil, fmt.Errorf("read token file: %w", err)
		}
		cfg.GitHub.GHToken = token
	}

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
		"insert", cfg.Insert.Enabled,
		"no-backup", cfg.Insert.NoBackup,
		"skip-header", cfg.SkipHeader,
	)
	ctlCfg := controller.Config{Files: cfg.Files, Serial: cfg.Serial}

	log.Info("App.New: init adapters ...")
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
	// bash gh-md-toc drops the path prefix only when a single document is requested.
	// InsertMd asks for bare anchors per document, so this stays a run-level rule and
	// remote documents keep their URL prefix even when --insert is set.
	rendererCfg.AbsolutePaths = len(cfg.Files) > 1
	renderer := coretoc.NewRenderer(rendererCfg)
	grabberRe := coretoc.NewGenerator(regexpExtractor, renderer)
	grabberJSON := coretoc.NewGenerator(jsonExtractor, renderer)
	getter := adapters.NewRemoteGetterWithClient(true, httpClient)
	temper := adapters.NewFileTemper()
	reader := adapters.NewFileReader()
	backupper := adapters.NewFileBackupper()
	stamper := adapters.NewStamper()

	log.Info("App.New: init usecases ...")
	var localChain markdownProcessor = localmd.New(cfg.Debug, checker, writer, converter, grabberRe, log)
	if cfg.SkipHeader {
		localChain = skipheader.New(localChain, reader, temper, log)
	}

	var ucLocal useCase = localChain
	if cfg.Insert.Enabled {
		ucLocal = insertmd.New(
			insertmd.Config{
				NoBackup:   cfg.Insert.NoBackup,
				HideFooter: cfg.Presentation.HideFooter,
			},
			localChain, reader, writer, backupper, stamper, notify, log,
		)
	}

	ucRemoteMD := remotemd.New(getter, localChain, temper, log)
	ucRemoteHTML := remotehtml.New(cfg.Debug, getter, temper, grabberJSON, log)

	log.Info("App.New: init controller ...")
	ctl := controller.New(ctlCfg, ucLocal, ucRemoteMD, ucRemoteHTML, log)

	log.Info("App.New: done.")
	return &App{
		ctl:    ctl,
		cfg:    cfg,
		notify: notify,
	}, nil
}
