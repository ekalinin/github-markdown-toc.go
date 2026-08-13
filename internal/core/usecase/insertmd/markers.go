package insertmd

import (
	"errors"
	"strings"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
)

var (
	ErrMarkersNotFound     = errors.New("no <!--ts--> / <!--te--> markers found")
	ErrMultipleMarkerPairs = errors.New("multiple <!--ts--> / <!--te--> marker pairs found")
	ErrMarkersOutOfOrder   = errors.New("the <!--te--> marker precedes <!--ts-->")
)

// replaceBetweenMarkers puts block between the TOC markers. The markers themselves
// and everything around them are left byte for byte as they were, so a document with
// CRLF endings keeps them outside the replaced block.
func replaceBetweenMarkers(content, block []byte) ([]byte, error) {
	lines := strings.Split(string(content), "\n")

	startIdx, endIdx := -1, -1
	starts, ends := 0, 0
	for i, line := range lines {
		switch strings.TrimSpace(line) {
		case entity.MarkerStart:
			starts++
			if startIdx < 0 {
				startIdx = i
			}
		case entity.MarkerEnd:
			ends++
			if endIdx < 0 {
				endIdx = i
			}
		}
	}

	switch {
	case starts == 0 || ends == 0:
		return nil, ErrMarkersNotFound
	case starts > 1 || ends > 1:
		return nil, ErrMultipleMarkerPairs
	case endIdx < startIdx:
		return nil, ErrMarkersOutOfOrder
	}

	result := make([]string, 0, len(lines))
	result = append(result, lines[:startIdx+1]...)
	if len(block) > 0 {
		result = append(result, strings.Split(string(block), "\n")...)
	}
	result = append(result, lines[endIdx:]...)
	return []byte(strings.Join(result, "\n")), nil
}
