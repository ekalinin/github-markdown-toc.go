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

type chunkExtractorStub struct {
	byInput map[string][]entity.Heading
}

func (s chunkExtractorStub) Extract(_ context.Context, input string) ([]entity.Heading, error) {
	return s.byInput[input], nil
}

func TestGeneratorRenumbersAnchorsAcrossChunks(t *testing.T) {
	extractor := chunkExtractorStub{byInput: map[string][]entity.Heading{
		"chunk-1": {{Level: 1, Text: "Usage", Anchor: "usage"}},
		"chunk-2": {
			{Level: 1, Text: "Usage", Anchor: "usage"},
			{Level: 1, Text: "Usage", Anchor: "usage-1"},
		},
	}}
	generator := NewGenerator(extractor, NewRenderer(DefaultConfig()))

	got, err := generator.Grab(context.Background(), "", "chunk-1", "chunk-2")
	if err != nil {
		t.Fatal(err)
	}
	want := entity.Toc{
		"* [Usage](#usage)",
		"* [Usage](#usage-1)",
		"* [Usage](#usage-2)",
	}
	if got == nil || !slices.Equal(*got, want) {
		t.Errorf("got TOC %v, want %v", got, want)
	}
}

// A single input is what GitHub already numbered for the whole document, so the
// anchors are passed through untouched, including a heading whose own text ends in a
// number.
func TestGeneratorKeepsAnchorsForASingleInput(t *testing.T) {
	extractor := extractorStub{headings: []entity.Heading{
		{Level: 1, Text: "Usage", Anchor: "usage"},
		{Level: 1, Text: "Usage 1", Anchor: "usage-1"},
	}}
	generator := NewGenerator(extractor, NewRenderer(DefaultConfig()))

	got, err := generator.Grab(context.Background(), "", "input")
	if err != nil {
		t.Fatal(err)
	}
	want := entity.Toc{
		"* [Usage](#usage)",
		"* [Usage 1](#usage-1)",
	}
	if got == nil || !slices.Equal(*got, want) {
		t.Errorf("got TOC %v, want %v", got, want)
	}
}
