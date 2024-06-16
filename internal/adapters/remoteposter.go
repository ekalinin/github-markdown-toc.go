package adapters

import "github.com/ekalinin/github-markdown-toc.go/internal/utils"

type RemotePoster struct {
}

func NewRemotePoster() *RemotePoster {
	return &RemotePoster{}
}

func (r *RemotePoster) Post(url, token, path string) (string, error) {
	return utils.HttpPost(url, path, token)
}
