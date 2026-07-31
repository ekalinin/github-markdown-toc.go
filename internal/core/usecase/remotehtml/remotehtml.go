package remotehtml

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"

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
	tempter ports.FileTemper
	log     ports.Logger
}

func New(cfg config.Config, getter ports.RemoteGetter,
	temper ports.FileTemper, grabber ports.TocGrabber, log ports.Logger) *RemoteHTML {
	return &RemoteHTML{cfg, getter, grabber, temper, log}
}

func (r *RemoteHTML) writeDebugFile(ctx context.Context, url string, body []byte) (path string, err error) {
	tmpfile, err := r.tempter.CreateTemp(ctx, "", "ghtoc-remote-*.debug.json")
	if err != nil {
		return "", fmt.Errorf("create debug file for %q: %w", url, err)
	}
	tempPath := tmpfile.Name()
	path = tempPath
	defer func() {
		if closeErr := tmpfile.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close debug file for %q: %w", url, closeErr))
		}
		if err != nil {
			if removeErr := r.tempter.Remove(tempPath); removeErr != nil {
				err = errors.Join(err, fmt.Errorf("remove debug file for %q: %w", url, removeErr))
			}
			path = ""
		}
	}()

	written, err := tmpfile.Write(body)
	if err != nil {
		return "", fmt.Errorf("write debug JSON for %q: %w", url, err)
	}
	if written != len(body) {
		return "", fmt.Errorf("write debug JSON for %q: %w", url, io.ErrShortWrite)
	}
	return path, nil
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

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, fmt.Errorf("parse content type for remote HTML %q: %w", url, err)
	}
	if mediaType != "application/json" {
		return nil, fmt.Errorf("get remote HTML %q: unexpected content type %q", url, contentType)
	}

	if r.cfg.Debug {
		debugFile, err := r.writeDebugFile(ctx, url, jsonBody)
		if err != nil {
			r.log.Info("RemoteHTML: writing json file failed", "err", err)
			return nil, err
		}
		r.log.Info("RemoteHTML: wrote debug JSON", "path", debugFile)
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
