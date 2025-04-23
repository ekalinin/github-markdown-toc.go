package adapters

import "github.com/ekalinin/github-markdown-toc.go/v2/internal/utils"

type RemoteGetter struct {
	asJSON bool
}

func NewRemoteGetter(asJSON bool) *RemoteGetter {
	return &RemoteGetter{asJSON: asJSON}
}

func (r *RemoteGetter) Get(path string) ([]byte, string, error) {
	if r.asJSON {
		return utils.HttpGetJson(path)
	}
	return utils.HttpGet(path)
}
