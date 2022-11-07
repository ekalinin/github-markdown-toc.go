package ghtoc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

const singleH1 = `
<h1><a id="user-content-document-title" class="anchor" aria-hidden="true" href="#document-title"><span aria-hidden="true" class="octicon octicon-link"></span></a>Document Title</h1>
`

const singleH2 = `
<h2>
  <a id="user-content-document-title" class="anchor" aria-hidden="true" href="#interesting-section">
    <span aria-hidden="true" class="octicon octicon-link"></span>
  </a>
  Interesting Section
</h2>
`

const multipleSections = `
<h1><a id="user-content-document-title" class="anchor" aria-hidden="true" href="#document-title"><span aria-hidden="true" class="octicon octicon-link"></span></a>Document Title</h1>
<span>Hi</span>
<h2><a id="user-content-document-title" class="anchor" aria-hidden="true" href="#first-section"><span aria-hidden="true" class="octicon octicon-link"></span></a>First Section</h2>
Some Text
<h3><a id="user-content-document-title" class="anchor" aria-hidden="true" href="#first-subsection"><span aria-hidden="true" class="octicon octicon-link"></span></a>First Subsection</h3>
<h2><a id="user-content-document-title" class="anchor" aria-hidden="true" href="#second-section"><span aria-hidden="true" class="octicon octicon-link"></span></a>Second Section</h2>
<h4><a id="user-content-document-title" class="anchor" aria-hidden="true" href="#second-subsection"><span aria-hidden="true" class="octicon octicon-link"></span></a>Second Subsection</h4>
`

func TestFindHeaders(t *testing.T) {
	t.Run("single H1", func(t *testing.T) {
		results := findHeadersInString(singleH1)
		assert.Len(t, results, 1)
		assert.Equal(
			t,
			Header{Depth: 0, Href: "#document-title", Name: "Document Title"},
			results[0],
		)
	})
	t.Run("single H2", func(t *testing.T) {
		results := findHeadersInString(singleH2)
		assert.Len(t, results, 1)
		assert.Equal(
			t,
			Header{Depth: 1, Href: "#interesting-section", Name: "Interesting Section"},
			results[0],
		)
	})
	t.Run("multiple sections", func(t *testing.T) {
		results := findHeadersInString(multipleSections)
		assert.Len(t, results, 5)
		assert.Equal(
			t,
			Header{Depth: 0, Href: "#document-title", Name: "Document Title"},
			results[0],
		)
		assert.Equal(
			t,
			Header{Depth: 1, Href: "#first-section", Name: "First Section"},
			results[1],
		)
		assert.Equal(
			t,
			Header{Depth: 2, Href: "#first-subsection", Name: "First Subsection"},
			results[2],
		)
		assert.Equal(
			t,
			Header{Depth: 1, Href: "#second-section", Name: "Second Section"},
			results[3],
		)
		assert.Equal(
			t,
			Header{Depth: 3, Href: "#second-subsection", Name: "Second Subsection"},
			results[4],
		)
	})
}

func TestFindAttribute(t *testing.T) {
	worldGreeting := html.Attribute{Namespace: "", Key: "greeting", Val: "Hello, World!"}
	spaceGreeting := html.Attribute{Namespace: "outer-space", Key: "greeting", Val: "Hello, Space!"}
	attrs := []html.Attribute{spaceGreeting, worldGreeting}
	t.Run("attribute exists", func(t *testing.T) {
		attr, ok := findAttribute(attrs, "", "greeting")
		assert.True(t, ok)
		assert.Equal(t, worldGreeting, attr)

		attr, ok = findAttribute(attrs, "outer-space", "greeting")
		assert.True(t, ok)
		assert.Equal(t, spaceGreeting, attr)
	})
	t.Run("attribute does not exist", func(t *testing.T) {
		t.Error("IMPLEMENT ME!")
	})
}
