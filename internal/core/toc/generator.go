package toc

import (
	"context"
	"fmt"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
)

// HeadingExtractor parses headings from an external document representation.
type HeadingExtractor interface {
	Extract(context.Context, string) ([]entity.Heading, error)
}

// Generator extracts headings and renders them as a Markdown TOC.
type Generator struct {
	extractor HeadingExtractor
	renderer  *Renderer
}

func NewGenerator(extractor HeadingExtractor, renderer *Renderer) *Generator {
	return &Generator{
		extractor: extractor,
		renderer:  renderer,
	}
}

// Grab extracts headings from every input and renders them as a single TOC for the
// document at path. The path is only used when the renderer is configured for
// absolute paths. Several inputs are the chunks of one document that was too large
// to convert in one request, and their anchors are renumbered as one document.
func (g *Generator) Grab(ctx context.Context, path string, inputs ...string) (*entity.Toc, error) {
	chunks := make([][]entity.Heading, 0, len(inputs))
	for _, input := range inputs {
		headings, err := g.extractor.Extract(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("extract headings: %w", err)
		}
		chunks = append(chunks, headings)
	}

	var headings []entity.Heading
	if len(chunks) == 1 {
		headings = chunks[0]
	} else {
		headings = renumberAnchors(chunks)
	}

	result, err := g.renderer.Render(ctx, path, headings)
	if err != nil {
		return nil, fmt.Errorf("render TOC: %w", err)
	}
	return &result, nil
}
