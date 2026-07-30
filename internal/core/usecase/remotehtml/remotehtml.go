package remotehtml

import (
	"context"
	"fmt"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/ports"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase/config"
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

func (r *RemoteHTML) Do(ctx context.Context, url string) (entity.Toc, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("process remote HTML %q: %w", url, err)
	}

	r.log.Info("RemoteHTML: start, downloading remote file ...", "url", url)
	jsonBody, contentType, err := r.getter.Get(ctx, url)
	if err != nil {
		r.log.Info("RemoteHTML: download fail", "err", err)
		return nil, fmt.Errorf("get remote HTML %q: %w", url, err)
	}
	r.log.Info("RemoteHTML: got file", "content-type=", contentType)

	if r.cfg.Debug {
		tmpfile, err := r.tempter.CreateTemp(ctx, "", "ghtoc-remote-json-*")
		if err != nil {
			r.log.Info("RemoteHTML: creating file failed", "err", err)
			return nil, fmt.Errorf("create debug file for %q: %w", url, err)
		}
		defer func() {
			if err := tmpfile.Close(); err != nil {
				r.log.Info("RemoteHTML: closing file failed", "err", err)
			}
		}()
		path := tmpfile.Name()

		jsonFile := path + ".debug.json"
		r.log.Info("RemoteHTML: writing json file", "path", jsonFile)
		if err := r.writer.Write(ctx, jsonFile, jsonBody); err != nil {
			r.log.Info("RemoteHTML: writing json file failed", "err", err)
			return nil, fmt.Errorf("write debug JSON for %q: %w", url, err)
		}
	}

	r.log.Info("RemoteHTML: grabbing the TOC ...")
	toc, err := r.grabber.Grab(ctx, string(jsonBody))
	if err != nil {
		r.log.Info("RemoteHTML: failed to grab TOC", "err", err)
		return nil, fmt.Errorf("grab TOC from remote HTML %q: %w", url, err)
	}

	r.log.Info("RemoteHTML: done.")
	if toc == nil {
		return entity.Toc{}, nil
	}
	return *toc, nil
}
