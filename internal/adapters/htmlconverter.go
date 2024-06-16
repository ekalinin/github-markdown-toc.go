package adapters

import (
	"github.com/ekalinin/github-markdown-toc.go/internal/core/ports"
)

type HTMLConverter struct {
	ghToken string
	ghURL   string
	poster  ports.RemotePoster
	log     ports.Logger
}

func NewHTMLConverter(token, url string, log ports.Logger) *HTMLConverter {
	return NewHTMLConverterX(token, url, NewRemotePoster(), log)
}

func NewHTMLConverterX(token, url string, poster ports.RemotePoster, log ports.Logger) *HTMLConverter {
	return &HTMLConverter{
		ghToken: token,
		ghURL:   url,
		poster:  poster,
		log:     log,
	}
}

func (c *HTMLConverter) Convert(file string) (string, error) {
	c.log.Info("adapters.HTMLConveter.Convert: start", "file", file)
	ghURL := c.ghURL + "/markdown/raw"
	c.log.Info("adapters.HTMLConveter.Convert: sending", "url", ghURL)
	return c.poster.Post(ghURL, c.ghToken, file)
}
