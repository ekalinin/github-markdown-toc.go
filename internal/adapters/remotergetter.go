package adapters

import (
	"context"
	"net/http"
)

type RemoteGetter struct {
	client *http.Client
	asJSON bool
}

func NewRemoteGetter(asJSON bool) *RemoteGetter {
	return NewRemoteGetterWithClient(asJSON, NewHTTPClient())
}

func NewRemoteGetterWithClient(asJSON bool, client *http.Client) *RemoteGetter {
	return &RemoteGetter{
		client: client,
		asJSON: asJSON,
	}
}

func (r *RemoteGetter) Get(ctx context.Context, path string) ([]byte, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}
	if r.asJSON {
		return HttpGetJson(ctx, r.client, path)
	}
	return HttpGet(ctx, r.client, path)
}
