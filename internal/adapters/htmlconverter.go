package adapters

import (
	"context"
	"net/http"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/ports"
)

type HTMLConverter struct {
	ghToken string
	ghURL   string
	poster  ports.RemotePoster
	log     ports.Logger
}

func NewHTMLConverter(token, url string, log ports.Logger) *HTMLConverter {
	return NewHTMLConverterWithClient(token, url, NewHTTPClient(), log)
}

func NewHTMLConverterWithClient(token, url string, client *http.Client, log ports.Logger) *HTMLConverter {
	return NewHTMLConverterX(token, url, NewRemotePosterWithClient(client), log)
}

func NewHTMLConverterX(token, url string, poster ports.RemotePoster, log ports.Logger) *HTMLConverter {
	return &HTMLConverter{
		ghToken: token,
		ghURL:   url,
		poster:  poster,
		log:     log,
	}
}

func (c *HTMLConverter) Convert(ctx context.Context, file string) (string, error) {
	c.log.Info("adapters.HTMLConverter.Convert: start", "file", file)
	ghURL := c.ghURL + "/markdown/raw"
	c.log.Info("adapters.HTMLConverter.Convert: sending", "url", ghURL)
	return c.poster.Post(ctx, ghURL, c.ghToken, file)
}
