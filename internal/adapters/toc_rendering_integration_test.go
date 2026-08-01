package adapters

import (
	"context"
	"slices"
	"testing"

	coretoc "github.com/ekalinin/github-markdown-toc.go/v2/internal/core/toc"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/version"
)

func TestJSONAndRegexpExtractorsRenderSameTOC(t *testing.T) {
	regexpExtractor, err := NewRegexpExtractor(version.GH_2024_03)
	if err != nil {
		t.Fatal(err)
	}
	renderer := coretoc.NewRenderer(coretoc.DefaultConfig())
	jsonGenerator := coretoc.NewGenerator(NewJSONExtractor(), renderer)
	regexpGenerator := coretoc.NewGenerator(regexpExtractor, renderer)

	jsonInput := `{"payload":{"blob":{"headerInfo":{"toc":[
		{"level":1,"text":"How to add a plugin?","anchor":"how-to-add-a-plugin"},
		{"level":2,"text":"Mandatory elements","anchor":"mandatory-elements"},
		{"level":3,"text":"The command plug_list_versions","anchor":"plug_list_versions"}
	]}}}}`
	jsonTOC, err := jsonGenerator.Grab(context.Background(), jsonInput)
	if err != nil {
		t.Fatal(err)
	}
	regexpTOC, err := regexpGenerator.Grab(context.Background(), htmlHeadingsV2024)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(*jsonTOC, *regexpTOC) {
		t.Errorf("JSON TOC %v does not match regexp TOC %v", *jsonTOC, *regexpTOC)
	}
}
