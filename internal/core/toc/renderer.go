package toc

import (
	"context"
	"strings"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
)

// Config controls how headings are filtered and rendered as a Markdown TOC.
type Config struct {
	Path          string
	AbsolutePaths bool
	StartDepth    int
	Depth         int
	Escape        bool
	Indent        int
}

func DefaultConfig() Config {
	return Config{
		Escape: true,
		Indent: 2,
	}
}

// Renderer converts parsed headings into a Markdown TOC.
type Renderer struct {
	cfg Config
}

func NewRenderer(cfg Config) *Renderer {
	return &Renderer{cfg: cfg}
}

func (r *Renderer) Render(ctx context.Context, headings []entity.Heading) (entity.Toc, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(headings) == 0 {
		return entity.Toc{}, nil
	}

	minLevel := headings[0].Level
	for _, heading := range headings[1:] {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		minLevel = min(minLevel, heading.Level)
	}
	baseLevel := max(minLevel, r.cfg.StartDepth+1)

	result := make(entity.Toc, 0, len(headings))
	for _, heading := range headings {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if heading.Level <= r.cfg.StartDepth {
			continue
		}
		if r.cfg.Depth > 0 && heading.Level > r.cfg.Depth {
			continue
		}

		text := heading.Text
		if r.cfg.Escape {
			text = escapeSpecialCharacters(text)
		}
		link := "#" + heading.Anchor
		if r.cfg.AbsolutePaths {
			link = r.cfg.Path + link
		}
		indent := strings.Repeat(" ", max(0, heading.Level-baseLevel)*max(0, r.cfg.Indent))
		result = append(result, indent+"* ["+text+"]("+link+")")
	}
	return result, nil
}

func escapeSpecialCharacters(value string) string {
	result := value
	for _, character := range []string{"\\", "`", "*", "_", "{", "}", "#", "+", "-", ".", "!"} {
		result = strings.ReplaceAll(result, character, "\\"+character)
	}
	return result
}
