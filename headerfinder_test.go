package ghtoc

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
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

// func assertHeaderEqual(t *testing.T, expected, actual Header) {
// 	if actual != expected {
// 		t.Errorf("Unexpected header value. actual: %+#v, expected: %+#v", actual, expected)
// 	}
// }

func TestFindHeaders(t *testing.T) {
	t.Run("single H1", func(t *testing.T) {
		// DEBUG BEGIN
		log.Printf("*** CHUCK: ===========")
		// DEBUG END
		results := findHeadersInString(singleH1)
		assert.Len(t, results, 1)
		assert.Equal(
			t,
			Header{Depth: 0, Href: "#document-title", Name: "Document Title"},
			results[0],
		)
	})
	t.Run("single H2", func(t *testing.T) {
		// DEBUG BEGIN
		log.Printf("*** CHUCK: ===========")
		// DEBUG END
		results := findHeadersInString(singleH2)
		assert.Len(t, results, 1)
		assert.Equal(
			t,
			Header{Depth: 1, Href: "#interesting-section", Name: "Interesting Section"},
			results[0],
		)
	})
}
