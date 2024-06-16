package remotemd

import (
	"strings"

	"github.com/ekalinin/github-markdown-toc.go/internal/core/entity"
	"github.com/ekalinin/github-markdown-toc.go/internal/core/ports"
	"github.com/ekalinin/github-markdown-toc.go/internal/core/usecase/config"
	"github.com/ekalinin/github-markdown-toc.go/internal/core/usecase/localmd"
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

func (r *RemoteMd) download(url string) (string, error) {
	body, ContentType, err := r.getter.Get(url)
	if err != nil {
		return "", err
	}

	// if not a plain text - it's an error
	if strings.Split(ContentType, ";")[0] != "text/plain" {
		r.log.Info("RemoteMD: not a plain text, stop.", "content-type", ContentType)
		return "", err
	}

	// if remote file's content is a plain text
	// we need to convert it to html
	tmpfile, err := r.temper.CreateTemp("", "ghtoc-remote-txt-*")
	if err != nil {
		r.log.Info("RemoteMD: creating tmp file failed.", "err", err)
		return "", err
	}
	defer tmpfile.Close()

	path := tmpfile.Name()
	r.log.Info("RemoteMD: save content into tmp file", "path", path)
	if err = r.writer.Write(tmpfile.Name(), body); err != nil {
		r.log.Info("RemoteMD: writing file failed.", "err", err)
		return "", err
	}
	return path, nil
}

func (r *RemoteMd) Do(url string) *entity.Toc {
	filename, err := r.download(url)
	if err != nil {
		r.log.Info("RemoteMD: download fail", "err", err)
		return nil
	}
	return r.ucLocalMD.Do(filename)
}
