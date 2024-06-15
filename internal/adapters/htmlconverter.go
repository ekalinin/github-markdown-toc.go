package adapters

import (
	"github.com/ekalinin/github-markdown-toc.go/internal/core/ports"
	"github.com/ekalinin/github-markdown-toc.go/internal/utils"
)

type HTMLConverter struct {
	ghToken string
	ghURL   string
	log     ports.Logger
}

func NewHTMLConverter(token, url string, log ports.Logger) *HTMLConverter {
	return &HTMLConverter{
		ghToken: token,
		ghURL:   url,
		log:     log,
	}
}

func (c *HTMLConverter) Convert(file string) (string, error) {
	c.log.Info("adapters.HTMLConveter.Convert: start", "file", file)
	ghURL := c.ghURL + "/markdown/raw"
	c.log.Info("adapters.HTMLConveter.Convert: sending", "url", ghURL)
	return utils.HttpPost(ghURL, file, c.ghToken)
}
