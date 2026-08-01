package adapters

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
)

func getTestJson() string {
	// how to get example:
	// ❯ curl -s -H 'Content-Type: application/json' -H 'Accept: application/json' \
	// 		https://github.com/ekalinin/sitemap.js/blob/6bc3eb12c898c1037a35a11b2eb24ababdeb3580/README.md | \
	// 		jq '.payload.blob.headerInfo.toc // .payload.codeViewBlobRoute.headerInfo.toc'
	// [
	// 	{
	// 	  "level": 1,
	// 	  "text": "sitemap.js",
	// 	  "anchor": "sitemapjs",
	// 	  "htmlText": "sitemap.js"
	// 	},
	// 	{
	// 	  "level": 2,
	// 	  "text": "Installation",
	// 	  "anchor": "installation",
	// 	  "htmlText": "Installation"
	// 	},
	// 	{
	// 	  "level": 2,
	// 	  "text": "Usage",
	// 	  "anchor": "usage",
	// 	  "htmlText": "Usage"
	// 	},
	// 	{
	// 	  "level": 2,
	// 	  "text": "License",
	// 	  "anchor": "license",
	// 	  "htmlText": "License"
	// 	}
	//   ]
	return `
	{
		"payload": {
			"blob": {
				"headerInfo": {
					"toc": [
						{
							"level": 1,
							"text": "sitemap.js",
							"anchor": "sitemapjs",
							"htmlText": "sitemap.js"
						},
						{
							"level": 2,
							"text": "Installation",
							"anchor": "installation",
							"htmlText": "Installation"
						},
						{
							"level": 2,
							"text": "Usage",
							"anchor": "usage",
							"htmlText": "Usage"
						},
						{
							"level": 3,
							"text": "Example",
							"anchor": "example",
							"htmlText": "Example"
						},
						{
							"level": 2,
							"text": "License",
							"anchor": "license",
							"htmlText": "License"
						}
					]
				}
			}
		}
	}
	`
}

func getTestJsonCodeViewRoute() string {
	return `
	{
		"payload": {
			"codeViewBlobRoute": {
				"headerInfo": {
					"toc": [
						{
							"level": 1,
							"text": "sitemap.js",
							"anchor": "sitemapjs",
							"htmlText": "sitemap.js"
						},
						{
							"level": 2,
							"text": "Installation",
							"anchor": "installation",
							"htmlText": "Installation"
						}
					]
				}
			}
		}
	}
	`
}

func TestJSONExtractorExtractsHeadings(t *testing.T) {
	extractor := NewJSONExtractor()
	got, err := extractor.Extract(context.Background(), getTestJson())
	if err != nil {
		t.Fatal(err)
	}
	want := []entity.Heading{
		{Level: 1, Text: "sitemap.js", Anchor: "sitemapjs"},
		{Level: 2, Text: "Installation", Anchor: "installation"},
		{Level: 2, Text: "Usage", Anchor: "usage"},
		{Level: 3, Text: "Example", Anchor: "example"},
		{Level: 2, Text: "License", Anchor: "license"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("got headings %v, want %v", got, want)
	}
}

func TestJSONExtractorUsesCodeViewBlobRoute(t *testing.T) {
	extractor := NewJSONExtractor()
	got, err := extractor.Extract(context.Background(), getTestJsonCodeViewRoute())
	if err != nil {
		t.Fatal(err)
	}
	want := []entity.Heading{
		{Level: 1, Text: "sitemap.js", Anchor: "sitemapjs"},
		{Level: 2, Text: "Installation", Anchor: "installation"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("got headings %v, want %v", got, want)
	}
}

func TestJSONExtractorNormalizesHeading(t *testing.T) {
	extractor := NewJSONExtractor()
	input := `{"payload":{"blob":{"headerInfo":{"toc":[{"level":2,"text":"  The command <code>foo</code>\n","anchor":"#the%20command"}]}}}}`
	got, err := extractor.Extract(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	want := []entity.Heading{{Level: 2, Text: "The command foo", Anchor: "the command"}}
	if !slices.Equal(got, want) {
		t.Errorf("got headings %v, want %v", got, want)
	}
}

func TestJSONExtractorRejectsMalformedJSON(t *testing.T) {
	_, err := NewJSONExtractor().Extract(context.Background(), `{`)
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestJSONExtractorReturnsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewJSONExtractor().Extract(ctx, getTestJson())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got error %v, want context cancellation", err)
	}
}
