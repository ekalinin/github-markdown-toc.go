package remotemd

import (
	"context"
	"fmt"
	"strings"

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
	writer    ports.FileWriter
	log       ports.Logger
}

func New(cfg config.Config, getter ports.RemoteGetter, localMD *localmd.LocalMd,
	temper ports.FileTemper, writer ports.FileWriter, log ports.Logger) *RemoteMd {
	return &RemoteMd{cfg, localMD, getter, temper, writer, log}
}

func (r *RemoteMd) download(ctx context.Context, url string) (string, error) {
	body, contentType, err := r.getter.Get(ctx, url)
	if err != nil {
		return "", fmt.Errorf("get remote Markdown %q: %w", url, err)
	}

	// if not a plain text - it's an error
	if strings.Split(contentType, ";")[0] != "text/plain" {
		r.log.Info("RemoteMD: not a plain text, stop.", "content-type", contentType)
		return "", fmt.Errorf("get remote Markdown %q: unexpected content type %q", url, contentType)
	}

	// if remote file's content is a plain text
	// we need to convert it to html
	tmpfile, err := r.temper.CreateTemp(ctx, "", "ghtoc-remote-txt-*")
	if err != nil {
		r.log.Info("RemoteMD: creating tmp file failed.", "err", err)
		return "", fmt.Errorf("create temporary file for %q: %w", url, err)
	}
	defer func() {
		if err := tmpfile.Close(); err != nil {
			r.log.Info("RemoteMD: closing file failed", "err", err)
		}
	}()

	path := tmpfile.Name()
	r.log.Info("RemoteMD: save content into tmp file", "path", path)
	if err = r.writer.Write(ctx, tmpfile.Name(), body); err != nil {
		r.log.Info("RemoteMD: writing file failed.", "err", err)
		return "", fmt.Errorf("write temporary file for %q: %w", url, err)
	}
	return path, nil
}

func (r *RemoteMd) Do(ctx context.Context, url string) (entity.Toc, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("process remote Markdown %q: %w", url, err)
	}

	filename, err := r.download(ctx, url)
	if err != nil {
		r.log.Info("RemoteMD: download fail", "err", err)
		return nil, err
	}
	toc, err := r.ucLocalMD.Do(ctx, filename)
	if err != nil {
		return nil, fmt.Errorf("process remote Markdown %q: %w", url, err)
	}
	return toc, nil
}
