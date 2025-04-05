package remotehtml

import (
	"github.com/ekalinin/github-markdown-toc.go/internal/core/entity"
	"github.com/ekalinin/github-markdown-toc.go/internal/core/ports"
	"github.com/ekalinin/github-markdown-toc.go/internal/core/usecase/config"
)

// - download json file
// - grab toc from json ()
type RemoteHTML struct {
	cfg     config.Config
	getter  ports.RemoteGetter
	grabber ports.TocGrabber
	writer  ports.FileWriter
	tempter ports.FileTemper
	log     ports.Logger
}

func New(cfg config.Config, getter ports.RemoteGetter, writer ports.FileWriter,
	temper ports.FileTemper, grabber ports.TocGrabber, log ports.Logger) *RemoteHTML {
	return &RemoteHTML{cfg, getter, grabber, writer, temper, log}
}

func (r *RemoteHTML) Do(url string) *entity.Toc {
	r.log.Info("RemoteHTML: start, downloading remote file ...", "url", url)
	jsonBody, ContentType, err := r.getter.Get(url)
	if err != nil {
		r.log.Info("RemoteHTML: download fail", "err", err)
		return nil
	}
	r.log.Info("RemoteHTML: got file", "content-type=", ContentType)

	if r.cfg.Debug {
		tmpfile, err := r.tempter.CreateTemp("", "ghtoc-remote-json-*")
		if err != nil {
			r.log.Info("RemoteHTML: creating file failed", "err", err)
			return nil
		}
		defer func() {
			if err := tmpfile.Close(); err != nil {
				r.log.Info("RemoteHTML: closing file failed", "err", err)
			}
		}()
		path := tmpfile.Name()

		jsonFile := path + ".debug.json"
		r.log.Info("RemoteHTML: writing json file", "path", jsonFile)
		if err := r.writer.Write(jsonFile, jsonBody); err != nil {
			r.log.Info("RemoteHTML: writing json file failed", "err", err)
			return nil
		}
	}

	r.log.Info("RemoteHTML: grabbing the TOC ...")
	toc, err := r.grabber.Grab(string(jsonBody))
	if err != nil {
		r.log.Info("RemoteHTML: failed to grab TOC", "err", err)
		return nil
	}

	r.log.Info("RemoteHTML: done.")
	return toc
}
