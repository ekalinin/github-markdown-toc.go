package adapters

import (
	"context"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/ports"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/utils"
)

type realPoster struct {
}

func (p *realPoster) Post(ctx context.Context, url, token, path string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
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

func (r *RemotePoster) Post(ctx context.Context, url, token, path string) (string, error) {
	return r.poster.Post(ctx, url, token, path)
}
