package remotemd

import (
	"github.com/ekalinin/github-markdown-toc.go/internal/core/entity"
	"github.com/ekalinin/github-markdown-toc.go/internal/core/usecase/config"
)

// - download remote file
// - call localmd use case
type RemoteMd struct {
	cfg config.Config
}

func New(cfg config.Config) *RemoteMd {
	return &RemoteMd{cfg}
}

func (r *RemoteMd) Do(file string) *entity.Toc {
	return nil
}
