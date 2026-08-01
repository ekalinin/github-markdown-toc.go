package adapters

import (
	"context"
	"net/http"
)

type RemotePoster struct {
	client *http.Client
}

func NewRemotePoster() *RemotePoster {
	return NewRemotePosterWithClient(NewHTTPClient())
}

func NewRemotePosterWithClient(client *http.Client) *RemotePoster {
	return &RemotePoster{client: client}
}

func (r *RemotePoster) Post(ctx context.Context, url, token, path string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	return HttpPost(ctx, r.client, url, path, token)
}
