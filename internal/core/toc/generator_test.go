package toc

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
)

type extractorStub struct {
	headings []entity.Heading
	err      error
}

func (s extractorStub) Extract(context.Context, string) ([]entity.Heading, error) {
	return s.headings, s.err
}

func TestGeneratorRendersExtractedHeadings(t *testing.T) {
	extractor := extractorStub{headings: []entity.Heading{{
		Level:  1,
		Text:   "Title",
		Anchor: "title",
	}}}
	generator := NewGenerator(extractor, NewRenderer(DefaultConfig()))

	got, err := generator.Grab(context.Background(), "", "input")
	if err != nil {
		t.Fatal(err)
	}
	want := entity.Toc{"* [Title](#title)"}
	if got == nil || !slices.Equal(*got, want) {
		t.Errorf("got TOC %v, want %v", got, want)
	}
}

func TestGeneratorPropagatesExtractorError(t *testing.T) {
	extractorErr := errors.New("extract failed")
	generator := NewGenerator(extractorStub{err: extractorErr}, NewRenderer(DefaultConfig()))

	_, err := generator.Grab(context.Background(), "", "input")
	if !errors.Is(err, extractorErr) {
		t.Fatalf("got error %v, want extractor error", err)
	}
}

func TestGeneratorPropagatesRendererCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	generator := NewGenerator(
		extractorStub{headings: []entity.Heading{{Level: 1}}},
		NewRenderer(DefaultConfig()),
	)

	_, err := generator.Grab(ctx, "", "input")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got error %v, want context cancellation", err)
	}
}

func TestGeneratorPassesPathToRenderer(t *testing.T) {
	extractor := extractorStub{headings: []entity.Heading{{
		Level:  1,
		Text:   "Title",
		Anchor: "title",
	}}}
	cfg := DefaultConfig()
	cfg.AbsolutePaths = true
	generator := NewGenerator(extractor, NewRenderer(cfg))

	got, err := generator.Grab(context.Background(), "docs/README.md", "input")
	if err != nil {
		t.Fatal(err)
	}
	want := entity.Toc{"* [Title](docs/README.md#title)"}
	if got == nil || !slices.Equal(*got, want) {
		t.Errorf("got TOC %v, want %v", got, want)
	}
}
