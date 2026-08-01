package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/utils"
)

// JSONExtractor extracts headings from a GitHub JSON document response.
type JSONExtractor struct{}

func NewJSONExtractor() *JSONExtractor {
	return &JSONExtractor{}
}

type jsonHeading struct {
	Level  int
	Text   string
	Anchor string
}

type blobWithTOC struct {
	HeaderInfo struct {
		TOC []jsonHeading `json:"toc"`
	} `json:"headerInfo"`
}

type tocWrapper struct {
	Payload struct {
		Blob              blobWithTOC
		CodeViewBlobRoute blobWithTOC `json:"codeViewBlobRoute"`
	}
}

func (w tocWrapper) tocSource() []jsonHeading {
	items := w.Payload.Blob.HeaderInfo.TOC
	if len(items) > 0 {
		return items
	}
	return w.Payload.CodeViewBlobRoute.HeaderInfo.TOC
}

func (*JSONExtractor) Extract(ctx context.Context, jsonBody string) ([]entity.Heading, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var wrapper tocWrapper
	if err := json.Unmarshal([]byte(jsonBody), &wrapper); err != nil {
		return nil, fmt.Errorf("unmarshal GitHub TOC JSON: %w", err)
	}

	items := wrapper.tocSource()
	headings := make([]entity.Heading, 0, len(items))
	for _, item := range items {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		anchor, err := url.QueryUnescape(item.Anchor)
		if err != nil {
			return nil, fmt.Errorf("unescape GitHub heading anchor %q: %w", item.Anchor, err)
		}
		headings = append(headings, entity.Heading{
			Level:  item.Level,
			Text:   utils.RemoveStuff(item.Text),
			Anchor: strings.TrimPrefix(anchor, "#"),
		})
	}
	return headings, nil
}
