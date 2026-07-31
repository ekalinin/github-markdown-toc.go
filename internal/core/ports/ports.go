package ports

import (
	"context"
	"os"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
)

type FileChecker interface {
	Exists(ctx context.Context, file string) (bool, error)
}

type FileWriter interface {
	Write(ctx context.Context, file string, data []byte) error
}

type HTMLConverter interface {
	Convert(ctx context.Context, file string) (string, error)
}

type TocGrabber interface {
	Grab(ctx context.Context, html string) (*entity.Toc, error)
}

type Logger interface {
	Info(format string, v ...any)
}

type RemoteGetter interface {
	Get(ctx context.Context, path string) ([]byte, string, error)
}

type FileTemper interface {
	CreateTemp(ctx context.Context, dir, pattern string) (*os.File, error)
	Remove(path string) error
}

type RemotePoster interface {
	Post(ctx context.Context, url, token, path string) (string, error)
}
