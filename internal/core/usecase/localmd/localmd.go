package localmd

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
)

type fileChecker interface {
	Exists(context.Context, string) (bool, error)
}

type fileWriter interface {
	Write(context.Context, string, []byte) error
}

type htmlConverter interface {
	Convert(context.Context, string) (string, error)
}

type tocGrabber interface {
	Grab(context.Context, string, string) (*entity.Toc, error)
}

type logger interface {
	Info(string, ...any)
}

// - read file
// - call gh api (md->html)
// - grab toc from html
type LocalMd struct {
	debug     bool
	checker   fileChecker
	writer    fileWriter
	converter htmlConverter
	grabber   tocGrabber

	log logger
}

func New(debug bool, checker fileChecker, writer fileWriter,
	converter htmlConverter, grabber tocGrabber, log logger) *LocalMd {
	return &LocalMd{
		debug:     debug,
		checker:   checker,
		writer:    writer,
		converter: converter,
		grabber:   grabber,
		log:       log,
	}
}

// Do builds the TOC for file, linking anchors relative to that same file.
func (uc *LocalMd) Do(ctx context.Context, file string) (entity.Toc, error) {
	return uc.DoAs(ctx, file, file)
}

// DoAs builds the TOC for file but renders links against displayPath. They differ
// when the content was downloaded or trimmed into a temporary file.
func (uc *LocalMd) DoAs(ctx context.Context, file, displayPath string) (entity.Toc, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("process local Markdown %q: %w", file, err)
	}

	uc.log.Info("LocalMD: Start", "file", file)
	exists, err := uc.checker.Exists(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("check local Markdown %q: %w", file, err)
	}
	if !exists {
		uc.log.Info("LocalMD: local file is not exists.")
		return nil, fmt.Errorf("open local Markdown %q: %w", file, fs.ErrNotExist)
	}

	uc.log.Info("LocalMD: converting to html ...")
	html, err := uc.converter.Convert(ctx, file)
	if err != nil {
		uc.log.Info("LocalMD: Failed to convert MD into HTML: %s", err)
		return nil, fmt.Errorf("convert local Markdown %q: %w", file, err)
	}

	if uc.debug {
		htmlFile := file + ".debug.html"
		uc.log.Info("LocalMD: writing html", "file", htmlFile)
		// TODO: move to port
		if err := uc.writer.Write(ctx, htmlFile, []byte(html)); err != nil {
			uc.log.Info("writing html file error: %s", err)
			return nil, fmt.Errorf("write debug HTML for %q: %w", file, err)
		}
	}

	uc.log.Info("LocalMD: grabbing the TOC ...")
	toc, err := uc.grabber.Grab(ctx, displayPath, html)
	if err != nil {
		uc.log.Info("LocalMD: failed to grab TOC: %s", err)
		return nil, fmt.Errorf("grab TOC from local Markdown %q: %w", file, err)
	}

	uc.log.Info("LocalMD: done.")
	if toc == nil {
		return entity.Toc{}, nil
	}
	return *toc, nil
}
