package adapters

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/mdsplit"
)

type remotePoster interface {
	Post(context.Context, string, string, string) (string, error)
	PostBody(context.Context, string, string, []byte) (string, error)
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

// maxChunkBytes is the largest payload the GitHub Markdown API renders. GitHub
// documents the limit as 400 KB; staying at 384 KB keeps us below it whether that
// means 400*1024 or 400000 bytes.
const maxChunkBytes = 384 << 10

// Convert renders file as HTML. The result holds one string per request: a single
// one for a document GitHub accepts whole, and one per chunk for a larger document.
func (c *HTMLConverter) Convert(ctx context.Context, file string) ([]string, error) {
	c.log.Info("adapters.HTMLConverter.Convert: start", "file", file)
	ghURL := c.ghURL + "/markdown/raw"
	c.log.Info("adapters.HTMLConverter.Convert: sending", "url", ghURL)

	if !exceedsChunkLimit(file) {
		html, err := c.poster.Post(ctx, ghURL, c.ghToken, file)
		if err != nil {
			return nil, withRateLimitHint(err)
		}
		return []string{html}, nil
	}
	return c.convertInChunks(ctx, ghURL, file)
}

// exceedsChunkLimit reports whether file is too large for one request. Only a
// successful stat sends us down the chunked path: when it fails, the single request
// reports the file error the way it always has.
func exceedsChunkLimit(file string) bool {
	info, err := os.Stat(file)
	return err == nil && info.Size() > maxChunkBytes
}

// convertInChunks converts a document GitHub would refuse as a whole. Requests are
// sent one after another: the order of the results is the order of the document, and
// a burst of parallel requests would only bring the rate limit closer.
func (c *HTMLConverter) convertInChunks(ctx context.Context, ghURL, file string) ([]string, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	chunks, err := mdsplit.Split(content, maxChunkBytes)
	if err != nil {
		return nil, err
	}
	c.log.Info("adapters.HTMLConverter.Convert: splitting large document",
		"file", file, "chunks", len(chunks))

	result := make([]string, 0, len(chunks))
	for i, chunk := range chunks {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		html, err := c.poster.PostBody(ctx, ghURL, c.ghToken, chunk)
		if err != nil {
			return nil, fmt.Errorf("convert chunk %d/%d: %w",
				i+1, len(chunks), withRateLimitHint(err))
		}
		result = append(result, html)
	}
	return result, nil
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
