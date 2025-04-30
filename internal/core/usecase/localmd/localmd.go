package localmd

import (
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/ports"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase/config"
)

// - read file
// - call gh api (md->html)
// - grab toc from html
type LocalMd struct {
	cfg       config.Config
	checker   ports.FileChecker
	writer    ports.FileWriter
	converter ports.HTMLConverter
	grabber   ports.TocGrabber

	log ports.Logger
}

func New(cfg config.Config, checker ports.FileChecker, writer ports.FileWriter,
	converter ports.HTMLConverter, grabber ports.TocGrabber, log ports.Logger) *LocalMd {
	return &LocalMd{
		cfg:       cfg,
		checker:   checker,
		writer:    writer,
		converter: converter,
		grabber:   grabber,
		log:       log,
	}
}

func (uc *LocalMd) Do(file string) *entity.Toc {
	uc.log.Info("LocalMD: Start", "file", file)
	if !uc.checker.Exists(file) {
		uc.log.Info("LocalMD: local file is not exists.")
		return nil
	}

	uc.log.Info("LocalMD: converting to html ...")
	html, err := uc.converter.Convert(file)
	if err != nil {
		uc.log.Info("LocalMD: Failed to convert MD into HTML: %s", err)
		return nil
	}

	if uc.cfg.Debug {
		htmlFile := file + ".debug.html"
		uc.log.Info("LocalMD: writing html", "file", htmlFile)
		// TODO: move to port
		if err := uc.writer.Write(htmlFile, []byte(html)); err != nil {
			uc.log.Info("writing html file error: %s", err)
			return nil
		}
	}

	uc.log.Info("LocalMD: grabbing the TOC ...")
	toc, err := uc.grabber.Grab(html)
	if err != nil {
		uc.log.Info("LocalMD: failed to grab TOC: %s", err)
		return nil
	}

	uc.log.Info("LocalMD: done.")
	return toc
}
