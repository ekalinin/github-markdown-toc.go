package remotehtml

import (
	"github.com/ekalinin/github-markdown-toc.go/internal/core/entity"
	"github.com/ekalinin/github-markdown-toc.go/internal/core/usecase/config"
)

// - download json file
// - grab toc from json ()
type RemoteHTML struct {
	cfg config.Config
}

func New(cfg config.Config) *RemoteHTML {
	return &RemoteHTML{cfg}
}

func (r *RemoteHTML) Do(file string) *entity.Toc {
	return nil
}
