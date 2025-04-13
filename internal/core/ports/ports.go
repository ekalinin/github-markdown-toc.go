package ports

import (
	"os"

	"github.com/ekalinin/github-markdown-toc.go/internal/core/entity"
)

type FileChecker interface {
	Exists(file string) bool
}

type FileWriter interface {
	Write(file string, data []byte) error
}

type HTMLConverter interface {
	Convert(file string) (string, error)
}

type TocGrabber interface {
	Grab(html string) (*entity.Toc, error)
}

type Logger interface {
	Info(format string, v ...any)
}

type RemoteGetter interface {
	Get(path string) ([]byte, string, error)
}

type FileTemper interface {
	CreateTemp(dir, pattern string) (*os.File, error)
}

type RemotePoster interface {
	Post(url, token, path string) (string, error)
}
