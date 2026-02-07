package adapters

import (
	"bytes"
	"fmt"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/utils"
)

const (
	MarkerStart = "<!--ts-->"
	MarkerEnd   = "<!--te-->"
)

type TocInserter struct {
	hideHeader bool
	hideFooter bool
}

func NewTocInserter(hideHeader, hideFooter bool) *TocInserter {
	return &TocInserter{
		hideHeader: hideHeader,
		hideFooter: hideFooter,
	}
}

func (ti *TocInserter) InsertToc(filepath string, toc string) error {
	stat, err := os.Stat(filepath)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	contentStr := string(content)

	if err := ti.validateMarkers(contentStr); err != nil {
		return fmt.Errorf("invalid markers in %s: %w", filepath, err)
	}

	var newContent string
	if ti.hasMarkers(contentStr) {
		newContent, err = ti.replaceContent(contentStr, toc)
		if err != nil {
			return err
		}
	} else {
		newContent = ti.insertAtTop(contentStr, toc)
	}

	return os.WriteFile(filepath, []byte(newContent), stat.Mode().Perm())
}

func (ti *TocInserter) hasMarkers(content string) bool {
	return strings.Contains(content, MarkerStart) && strings.Contains(content, MarkerEnd)
}

func (ti *TocInserter) validateMarkers(content string) error {
	startCount := strings.Count(content, MarkerStart)
	endCount := strings.Count(content, MarkerEnd)

	if startCount != endCount {
		return fmt.Errorf("mismatched markers: found %d start markers and %d end markers", startCount, endCount)
	}
	if startCount > 1 {
		return fmt.Errorf("multiple marker pairs found (%d pairs), only one pair is supported", startCount)
	}
	if startCount == 0 && endCount == 0 {
		return nil
	}
	return nil
}

func (ti *TocInserter) replaceContent(content, toc string) (string, error) {
	lines := strings.Split(content, "\n")
	var result []string
	var insideMarkers bool
	var markerFound bool

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == MarkerStart {
			result = append(result, line)
			insideMarkers = true
			markerFound = true
			result = append(result, ti.formatToc(toc))
			continue
		}

		if trimmed == MarkerEnd {
			result = append(result, line)
			insideMarkers = false
			continue
		}

		if !insideMarkers {
			result = append(result, line)
		}
	}

	if !markerFound {
		return "", fmt.Errorf("markers not found")
	}

	return strings.Join(result, "\n"), nil
}

func (ti *TocInserter) formatToc(toc string) string {
	var buf bytes.Buffer

	if !ti.hideHeader {
		buf.WriteString(utils.GetHeaderText())
	}

	buf.WriteString("\n")
	buf.WriteString(toc)

	if !ti.hideFooter {
		buf.WriteString("\n")
		buf.WriteString(ti.generateTimestamp())
		buf.WriteString("\n")
		buf.WriteString(utils.GetFooterText())
		buf.WriteString("\n")
	}

	return buf.String()
}

func (ti *TocInserter) generateTimestamp() string {
	username := "unknown"
	if u, err := user.Current(); err == nil {
		username = u.Username
	}

	timestamp := time.Now().Format("2006-01-02T15:04-07:00")
	return fmt.Sprintf("<!-- Added by: %s, at: %s -->", username, timestamp)
}

func (ti *TocInserter) insertAtTop(content, toc string) string {
	lines := strings.Split(content, "\n")
	var result []string
	var insertIndex int
	var foundHeading bool

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			insertIndex = i
			foundHeading = true
			break
		}
	}

	if !foundHeading {
		insertIndex = 0
	}

	result = append(result, lines[:insertIndex]...)
	result = append(result, MarkerStart)
	result = append(result, ti.formatToc(toc))
	result = append(result, MarkerEnd)
	result = append(result, "")
	result = append(result, lines[insertIndex:]...)

	return strings.Join(result, "\n")
}
