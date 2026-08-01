package toc

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
)

func TestRendererRender(t *testing.T) {
	headings := []entity.Heading{
		{Level: 1, Text: "Root.", Anchor: "root"},
		{Level: 2, Text: "Child_*", Anchor: "child_"},
		{Level: 3, Text: "Grand!", Anchor: "grand"},
		{Level: 2, Text: "Second", Anchor: "second"},
	}
	tests := []struct {
		name     string
		cfg      Config
		headings []entity.Heading
		want     entity.Toc
	}{
		{
			name: "default",
			cfg:  DefaultConfig(),
			want: entity.Toc{
				"* [Root\\.](#root)",
				"  * [Child\\_\\*](#child_)",
				"    * [Grand\\!](#grand)",
				"  * [Second](#second)",
			},
		},
		{
			name: "depth",
			cfg:  Config{Depth: 2, Escape: true, Indent: 2},
			want: entity.Toc{
				"* [Root\\.](#root)",
				"  * [Child\\_\\*](#child_)",
				"  * [Second](#second)",
			},
		},
		{
			name: "start depth",
			cfg:  Config{StartDepth: 1, Escape: true, Indent: 2},
			want: entity.Toc{
				"* [Child\\_\\*](#child_)",
				"  * [Grand\\!](#grand)",
				"* [Second](#second)",
			},
		},
		{
			name: "custom indentation",
			cfg:  Config{Escape: true, Indent: 4},
			want: entity.Toc{
				"* [Root\\.](#root)",
				"    * [Child\\_\\*](#child_)",
				"        * [Grand\\!](#grand)",
				"    * [Second](#second)",
			},
		},
		{
			name: "absolute paths",
			cfg: Config{
				Path:          "README.md",
				AbsolutePaths: true,
				Escape:        true,
				Indent:        2,
			},
			want: entity.Toc{
				"* [Root\\.](README.md#root)",
				"  * [Child\\_\\*](README.md#child_)",
				"    * [Grand\\!](README.md#grand)",
				"  * [Second](README.md#second)",
			},
		},
		{
			name:     "escaping",
			cfg:      DefaultConfig(),
			headings: []entity.Heading{{Level: 1, Text: `abc\*_{}#+-.!`, Anchor: "special"}},
			want:     entity.Toc{`* [abc\\\*\_\{\}\#\+\-\.\!](#special)`},
		},
		{
			name:     "backtick escaping",
			cfg:      DefaultConfig(),
			headings: []entity.Heading{{Level: 1, Text: "`code`", Anchor: "code"}},
			want:     entity.Toc{"* [\\`code\\`](#code)"},
		},
		{
			name: "escaping disabled",
			cfg:  Config{Indent: 2},
			want: entity.Toc{
				"* [Root.](#root)",
				"  * [Child_*](#child_)",
				"    * [Grand!](#grand)",
				"  * [Second](#second)",
			},
		},
		{
			name: "start depth without level one",
			cfg:  Config{StartDepth: 1, Escape: true, Indent: 2},
			headings: []entity.Heading{
				{Level: 2, Text: "Child", Anchor: "child"},
				{Level: 3, Text: "Grand", Anchor: "grand"},
			},
			want: entity.Toc{
				"* [Child](#child)",
				"  * [Grand](#grand)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := tt.headings
			if input == nil {
				input = headings
			}
			got, err := NewRenderer(tt.cfg).Render(context.Background(), input)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("got TOC %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRendererReturnsEmptyTOC(t *testing.T) {
	got, err := NewRenderer(DefaultConfig()).Render(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("got TOC %v, want a non-nil empty TOC", got)
	}
}

func TestRendererReturnsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewRenderer(DefaultConfig()).Render(ctx, []entity.Heading{{Level: 1}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got error %v, want context cancellation", err)
	}
}
