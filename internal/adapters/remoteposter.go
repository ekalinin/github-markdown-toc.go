package adapters

import (
	"context"
	"net/http"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/ports"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/utils"
)

type realPoster struct {
	client *http.Client
}

func (p *realPoster) Post(ctx context.Context, url, token, path string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return utils.HttpPost(ctx, p.client, url, path, token)
}

type RemotePoster struct {
	poster ports.RemotePoster
}

func NewRemotePoster() *RemotePoster {
	return NewRemotePosterWithClient(NewHTTPClient())
}

func NewRemotePosterWithClient(client *http.Client) *RemotePoster {
	return NewRemotePosterX(&realPoster{client: client})
}

func NewRemotePosterX(poster ports.RemotePoster) *RemotePoster {
	return &RemotePoster{poster: poster}
}

func (r *RemotePoster) Post(ctx context.Context, url, token, path string) (string, error) {
	return r.poster.Post(ctx, url, token, path)
}
