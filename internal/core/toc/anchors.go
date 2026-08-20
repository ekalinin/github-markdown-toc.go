package toc

import (
	"strconv"
	"strings"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
)

// renumberAnchors makes anchors unique across chunks that GitHub numbered
// separately. GitHub appends "-N" to the Nth repeat of a slug within one rendered
// document, so a document converted in pieces restarts that numbering in every
// piece. The base slug is recovered with the same counter GitHub used inside the
// chunk, then numbered again against a counter that spans the whole document.
func renumberAnchors(chunks [][]entity.Heading) []entity.Heading {
	total := 0
	for _, chunk := range chunks {
		total += len(chunk)
	}

	result := make([]entity.Heading, 0, total)
	global := make(map[string]int)
	for _, chunk := range chunks {
		local := make(map[string]int)
		for _, heading := range chunk {
			base := baseAnchor(heading.Anchor, local)
			local[base]++
			if n := global[base]; n > 0 {
				heading.Anchor = base + "-" + strconv.Itoa(n)
			} else {
				heading.Anchor = base
			}
			global[base]++
			result = append(result, heading)
		}
	}
	return result
}

// baseAnchor strips the duplicate counter GitHub appended inside a chunk. A suffix
// only counts as a counter when it matches the number of times the base was already
// seen in that same chunk, which is exactly when GitHub would have produced it.
func baseAnchor(anchor string, local map[string]int) string {
	idx := strings.LastIndexByte(anchor, '-')
	if idx <= 0 || idx == len(anchor)-1 {
		return anchor
	}
	suffix := anchor[idx+1:]
	n, err := strconv.Atoi(suffix)
	if err != nil || n <= 0 || suffix != strconv.Itoa(n) {
		return anchor
	}
	base := anchor[:idx]
	if local[base] != n {
		return anchor
	}
	return base
}
