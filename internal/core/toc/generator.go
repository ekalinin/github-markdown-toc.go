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

// Grab extracts headings from input and renders them as a TOC for the document at
// path. The path is only used when the renderer is configured for absolute paths.
func (g *Generator) Grab(ctx context.Context, path, input string) (*entity.Toc, error) {
	headings, err := g.extractor.Extract(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("extract headings: %w", err)
	}

	result, err := g.renderer.Render(ctx, path, headings)
	if err != nil {
		return nil, fmt.Errorf("render TOC: %w", err)
	}
	return &result, nil
}
