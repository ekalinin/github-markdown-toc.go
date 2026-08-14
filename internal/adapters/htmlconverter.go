package adapters

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

type remotePoster interface {
	Post(context.Context, string, string, string) (string, error)
}

type HTMLConverter struct {
	ghToken string
	ghURL   string
	poster  remotePoster
	log     logger
}

func NewHTMLConverter(token, url string, log logger) *HTMLConverter {
	return NewHTMLConverterWithClient(token, url, NewHTTPClient(), log)
}

func NewHTMLConverterWithClient(token, url string, client *http.Client, log logger) *HTMLConverter {
	return NewHTMLConverterX(token, url, NewRemotePosterWithClient(client), log)
}

func NewHTMLConverterX(token, url string, poster remotePoster, log logger) *HTMLConverter {
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

	html, err := c.poster.Post(ctx, ghURL, c.ghToken, file)
	if err != nil {
		return "", withRateLimitHint(err)
	}
	return html, nil
}

// withRateLimitHint points the user at the token options when GitHub throttles us.
// Without a token the markdown endpoint allows very few requests per hour.
func withRateLimitHint(err error) error {
	var statusErr *HTTPStatusError
	if !errors.As(err, &statusErr) {
		return err
	}
	if statusErr.StatusCode != http.StatusForbidden &&
		statusErr.StatusCode != http.StatusTooManyRequests {
		return err
	}
	return fmt.Errorf("%w (GitHub API rate limit reached, pass --token or set GH_TOC_TOKEN)", err)
}
