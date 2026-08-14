package adapters

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
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

// rateLimitMarker is what GitHub puts in the body when the API rate limit is hit.
// The bash gh-md-toc greps the response for the same string.
const rateLimitMarker = "API rate limit exceeded"

// withRateLimitHint points the user at the token options when GitHub throttles us.
// A 403 on its own is not enough: GitHub also returns it for bad credentials and for
// a token missing a scope, and sending those users off to fetch a token would be
// misleading. A 429 is unambiguous.
func withRateLimitHint(err error) error {
	var statusErr *HTTPStatusError
	if !errors.As(err, &statusErr) {
		return err
	}

	rateLimited := statusErr.StatusCode == http.StatusTooManyRequests ||
		(statusErr.StatusCode == http.StatusForbidden &&
			strings.Contains(statusErr.Body, rateLimitMarker))
	if !rateLimited {
		return err
	}
	return fmt.Errorf(
		"%w (GitHub API rate limit reached, pass --token, set GH_TOC_TOKEN, "+
			"or put the token in token.txt next to the binary)", err)
}
