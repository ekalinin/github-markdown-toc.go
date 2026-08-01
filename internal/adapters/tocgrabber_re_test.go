package adapters

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/version"
)

const htmlHeadingV0 = `
<h1><a id="user-content-readme-in-another-language" class="anchor" href="#readme-in-another-language" aria-hidden="true"><span class="octicon octicon-link"></span></a>README in another language</h1>`

const htmlHeadingsV2023 = `
<h1 id="user-content-how-to-add-a-plugin"><a class="heading-link" href="#how-to-add-a-plugin">How to add a plugin?<span aria-hidden="true" class="octicon octicon-link"></span></a></h1>
<h2 id="user-content-mandatory-elements"><a class="heading-link" href="#mandatory-elements">Mandatory elements<span aria-hidden="true" class="octicon octicon-link"></span></a></h2>
<h3 id="user-content-plug_list_versions"><a class="heading-link" href="#plug_list_versions">The command <code>plug_list_versions</code>
<span aria-hidden="true" class="octicon octicon-link"></span></a></h3>`

const htmlHeadingsV2024 = `
<div class="markdown-heading"><h1 class="heading-element">How to add a plugin?</h1><a id="user-content-how-to-add-a-plugin" class="anchor-element" aria-label="Permalink: How to add a plugin?" href="#how-to-add-a-plugin"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
<div class="markdown-heading"><h2 class="heading-element">Mandatory elements</h2><a id="user-content-mandatory-elements" class="anchor-element" aria-label="Permalink: Mandatory elements" href="#mandatory-elements"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>
<div class="markdown-heading"><h3 class="heading-element">The command <code>plug_list_versions</code>
</h3><a id="user-content-plug_list_versions" class="anchor-element" aria-label="Permalink: plug_list_versions" href="#plug_list_versions"><span aria-hidden="true" class="octicon octicon-link"></span></a></div>`

func TestRegexpExtractorExtractsLegacyHeading(t *testing.T) {
	extractor, err := NewRegexpExtractor(version.GH_V0)
	if err != nil {
		t.Fatal(err)
	}
	got, err := extractor.Extract(context.Background(), htmlHeadingV0)
	if err != nil {
		t.Fatal(err)
	}
	want := []entity.Heading{{
		Level:  1,
		Text:   "README in another language",
		Anchor: "readme-in-another-language",
	}}
	if !slices.Equal(got, want) {
		t.Errorf("got headings %v, want %v", got, want)
	}
}

func TestRegexpExtractorExtractsCurrentHeadings(t *testing.T) {
	want := []entity.Heading{
		{Level: 1, Text: "How to add a plugin?", Anchor: "how-to-add-a-plugin"},
		{Level: 2, Text: "Mandatory elements", Anchor: "mandatory-elements"},
		{Level: 3, Text: "The command plug_list_versions", Anchor: "plug_list_versions"},
	}
	tests := []struct {
		name    string
		version string
		html    string
	}{
		{name: "2023-10", version: version.GH_2023_10, html: htmlHeadingsV2023},
		{name: "2024-03", version: version.GH_2024_03, html: htmlHeadingsV2024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor, err := NewRegexpExtractor(tt.version)
			if err != nil {
				t.Fatal(err)
			}
			got, err := extractor.Extract(context.Background(), tt.html)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got, want) {
				t.Errorf("got headings %v, want %v", got, want)
			}
		})
	}
}

func TestRegexpExtractorUnescapesAnchor(t *testing.T) {
	extractor, err := NewRegexpExtractor(version.GH_2024_03)
	if err != nil {
		t.Fatal(err)
	}
	html := `<div class="markdown-heading"><h2 class="heading-element">Hello</h2><a id="user-content-hello" class="anchor-element" aria-label="Permalink: Hello" href="#hello%20world"></a></div>`
	got, err := extractor.Extract(context.Background(), html)
	if err != nil {
		t.Fatal(err)
	}
	want := []entity.Heading{{Level: 2, Text: "Hello", Anchor: "hello world"}}
	if !slices.Equal(got, want) {
		t.Errorf("got headings %v, want %v", got, want)
	}
}

func TestNewRegexpExtractorRejectsUnknownVersion(t *testing.T) {
	_, err := NewRegexpExtractor("unknown")
	if err == nil {
		t.Fatal("expected an error for an unsupported regexp version")
	}
	want := "unsupported GitHub regexp version \"unknown\"; supported versions: 0, 2023-10, 2024-03"
	if err.Error() != want {
		t.Errorf("got error %q, want %q", err, want)
	}
}

func TestRegexpExtractorReturnsContextCancellation(t *testing.T) {
	extractor, err := NewRegexpExtractor(version.GH_2024_03)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = extractor.Extract(ctx, htmlHeadingsV2024)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got error %v, want context cancellation", err)
	}
}
