package ghtoc

import "testing"

const singleHdr = `
<h1><a id="user-content-document-title" class="anchor" aria-hidden="true" href="#document-title"><span aria-hidden="true" class="octicon octicon-link"></span></a>Document Title</h1>
`

func TestHeaderRegexp(t *testing.T) {
	r := newHeaderRegexp()

	results := r.FindAllStringSubmatch(singleHdr, -1)
	if len(results) != 1 {
		t.Errorf("Expected a single header. %+#v", results)
	}
}
