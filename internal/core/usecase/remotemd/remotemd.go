package remotemd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/ports"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase/config"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase/localmd"
)

// - download remote file
// - call localmd use case
type RemoteMd struct {
	cfg       config.Config
	ucLocalMD *localmd.LocalMd
	getter    ports.RemoteGetter
	temper    ports.FileTemper
	log       ports.Logger
}

func New(cfg config.Config, getter ports.RemoteGetter, localMD *localmd.LocalMd,
	temper ports.FileTemper, log ports.Logger) *RemoteMd {
	return &RemoteMd{cfg, localMD, getter, temper, log}
}

func (r *RemoteMd) download(ctx context.Context, url string) (path string, err error) {
	body, contentType, err := r.getter.Get(ctx, url)
	if err != nil {
		return "", fmt.Errorf("get remote Markdown %q: %w", url, err)
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", fmt.Errorf("parse content type for remote Markdown %q: %w", url, err)
	}
	if mediaType != "text/plain" {
		r.log.Info("RemoteMD: not a plain text, stop.", "content-type", contentType)
		return "", fmt.Errorf("get remote Markdown %q: unexpected content type %q", url, contentType)
	}

	tmpfile, err := r.temper.CreateTemp(ctx, "", "ghtoc-remote-txt-*")
	if err != nil {
		r.log.Info("RemoteMD: creating tmp file failed.", "err", err)
		return "", fmt.Errorf("create temporary file for %q: %w", url, err)
	}
	tempPath := tmpfile.Name()
	path = tempPath
	defer func() {
		if closeErr := tmpfile.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close temporary file for %q: %w", url, closeErr))
		}
		if err != nil {
			if removeErr := r.temper.Remove(tempPath); removeErr != nil {
				err = errors.Join(err, fmt.Errorf("remove temporary file for %q: %w", url, removeErr))
			}
			path = ""
		}
	}()

	r.log.Info("RemoteMD: save content into tmp file", "path", path)
	written, err := tmpfile.Write(body)
	if err != nil {
		r.log.Info("RemoteMD: writing file failed.", "err", err)
		return "", fmt.Errorf("write temporary file for %q: %w", url, err)
	}
	if written != len(body) {
		return "", fmt.Errorf("write temporary file for %q: %w", url, io.ErrShortWrite)
	}
	return path, nil
}

func (r *RemoteMd) Do(ctx context.Context, url string) (toc entity.Toc, err error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("process remote Markdown %q: %w", url, err)
	}

	filename, err := r.download(ctx, url)
	if err != nil {
		r.log.Info("RemoteMD: download fail", "err", err)
		return nil, err
	}
	defer func() {
		if removeErr := r.temper.Remove(filename); removeErr != nil {
			err = errors.Join(err, fmt.Errorf("remove temporary file for remote Markdown %q: %w", url, removeErr))
		}
	}()

	toc, err = r.ucLocalMD.Do(ctx, filename)
	if err != nil {
		return nil, fmt.Errorf("process remote Markdown %q: %w", url, err)
	}
	return toc, nil
}
