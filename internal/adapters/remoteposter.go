package adapters

import (
	"github.com/ekalinin/github-markdown-toc.go/internal/core/ports"
	"github.com/ekalinin/github-markdown-toc.go/internal/utils"
)

type realPoster struct {
}

func (p *realPoster) Post(url, token, path string) (string, error) {
	return utils.HttpPost(url, path, token)
}

type RemotePoster struct {
	poster ports.RemotePoster
}

func NewRemotePoster() *RemotePoster {
	return NewRemotePosterX(&realPoster{})
}

func NewRemotePosterX(poster ports.RemotePoster) *RemotePoster {
	return &RemotePoster{poster: poster}
}

func (r *RemotePoster) Post(url, token, path string) (string, error) {
	return r.poster.Post(url, token, path)
}
