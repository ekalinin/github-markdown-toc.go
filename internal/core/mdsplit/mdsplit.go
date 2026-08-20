// Package mdsplit cuts a Markdown document into pieces small enough for GitHub's
// Markdown API, which refuses to render a payload larger than 400 KB.
package mdsplit

import (
	"bytes"
	"fmt"
	"strings"
)

// blockKind is the kind of block a line can be inside of. Only blocks that survive a
// blank line are tracked: everything else ends at the blank line we cut on anyway.
type blockKind int

const (
	blockNone blockKind = iota
	blockFence
	blockComment
	blockRawHTML
)

// rawHTMLTags are the tags whose content GitHub keeps verbatim across blank lines.
var rawHTMLTags = []string{"pre", "script", "style", "textarea"}

// state is the block context at the end of the lines processed so far.
type state struct {
	kind      blockKind
	openLine  string
	fenceChar byte
	fenceLen  int
	tag       string
}

// closing is the line that ends the open block at the end of a chunk.
func (s state) closing() string {
	switch s.kind {
	case blockFence:
		return strings.Repeat(string(s.fenceChar), s.fenceLen) + "\n"
	case blockComment:
		return "-->\n"
	case blockRawHTML:
		return "</" + s.tag + ">\n"
	}
	return ""
}

// reopening is the line that restores the block at the start of the next chunk.
func (s state) reopening() string {
	switch s.kind {
	case blockFence:
		if strings.HasSuffix(s.openLine, "\n") {
			return s.openLine
		}
		return s.openLine + "\n"
	case blockComment:
		return "<!--\n"
	case blockRawHTML:
		return "<" + s.tag + ">\n"
	}
	return ""
}

// Split cuts content into chunks of at most maxBytes bytes.
//
// A chunk normally ends at a blank line outside any fenced code block, HTML comment
// or raw HTML block. That is where GitHub ends a block too, so every chunk is parsed
// the way the same lines would be parsed inside the whole document. When no such
// point fits into maxBytes, the cut falls on a line boundary and the open block is
// closed at the end of the chunk and reopened at the start of the next one, which
// keeps code and comments from turning into Markdown.
func Split(content []byte, maxBytes int) ([][]byte, error) {
	if maxBytes <= 0 {
		return nil, fmt.Errorf("mdsplit: chunk limit must be positive, got %d", maxBytes)
	}

	lines := splitLines(content)
	if len(lines) == 0 {
		return [][]byte{}, nil
	}

	var (
		chunks  [][]byte
		cur     []byte
		st      state
		safeLen int
	)

	for i, line := range lines {
		next := advance(st, line)

		// Preferred cut: the last blank line outside a block.
		if len(cur) > 0 && safeLen > 0 && !fits(len(cur), line, next, maxBytes) {
			chunks = append(chunks, bytes.Clone(cur[:safeLen]))
			cur = bytes.Clone(cur[safeLen:])
			safeLen = 0
		}

		// Forced cut: no blank line fits, so cut on a line boundary instead.
		if len(cur) > 0 && !fits(len(cur), line, next, maxBytes) {
			cut := len(cur)
			// Keep a setext heading whole: cutting between the text and its
			// underline would turn the heading into a horizontal rule.
			if st.kind == blockNone && isSetextUnderline(line) {
				if start := lastLineStart(cur); start > 0 {
					cut = start
				}
			}
			chunk := bytes.Clone(cur[:cut])
			if cut == len(cur) {
				chunk = append(chunk, st.closing()...)
			}
			chunks = append(chunks, chunk)
			cur = append([]byte(st.reopening()), cur[cut:]...)
			safeLen = 0
		}

		if !fits(len(cur), line, next, maxBytes) {
			return nil, fmt.Errorf(
				"mdsplit: line %d is %d bytes and does not fit into a %d byte chunk; "+
					"split the document manually", i+1, len(line), maxBytes)
		}

		cur = append(cur, line...)
		st = next
		if st.kind == blockNone && isBlank(line) {
			safeLen = len(cur)
		}
	}

	if len(cur) > 0 {
		chunks = append(chunks, cur)
	}
	return chunks, nil
}

// fits reports whether line still leaves room for the closing line the chunk would
// need if it ended right after that line.
func fits(curLen int, line string, next state, maxBytes int) bool {
	return curLen+len(line)+len(next.closing()) <= maxBytes
}

// advance returns the block context after line.
func advance(st state, line string) state {
	switch st.kind {
	case blockFence:
		if closesFence(line, st.fenceChar, st.fenceLen) {
			return state{}
		}
		return st
	case blockComment:
		if strings.Contains(line, "-->") {
			return state{}
		}
		return st
	case blockRawHTML:
		if strings.Contains(strings.ToLower(line), "</"+st.tag+">") {
			return state{}
		}
		return st
	}

	if char, length, ok := opensFence(line); ok {
		return state{kind: blockFence, openLine: line, fenceChar: char, fenceLen: length}
	}
	if rest, ok := opensComment(line); ok {
		if strings.Contains(rest, "-->") {
			return state{}
		}
		return state{kind: blockComment}
	}
	if tag, ok := opensRawHTML(line); ok {
		if strings.Contains(strings.ToLower(line), "</"+tag+">") {
			return state{}
		}
		return state{kind: blockRawHTML, tag: tag}
	}
	return state{}
}

// splitLines cuts content into lines that keep their trailing newline.
func splitLines(content []byte) []string {
	if len(content) == 0 {
		return nil
	}
	lines := strings.SplitAfter(string(content), "\n")
	if last := len(lines) - 1; lines[last] == "" {
		lines = lines[:last]
	}
	return lines
}

// trimIndent strips leading spaces and reports how many there were.
func trimIndent(line string) (string, int) {
	i := 0
	for i < len(line) && line[i] == ' ' {
		i++
	}
	return line[i:], i
}

func isBlank(line string) bool {
	return strings.TrimSpace(line) == ""
}

// opensFence reports the fence character and length when line opens a fenced block.
func opensFence(line string) (byte, int, bool) {
	trimmed, indent := trimIndent(line)
	if indent > 3 || trimmed == "" {
		return 0, 0, false
	}
	char := trimmed[0]
	if char != '`' && char != '~' {
		return 0, 0, false
	}
	length := 0
	for length < len(trimmed) && trimmed[length] == char {
		length++
	}
	if length < 3 {
		return 0, 0, false
	}
	// A backtick fence cannot carry a backtick in its info string.
	if char == '`' && strings.Contains(trimmed[length:], "`") {
		return 0, 0, false
	}
	return char, length, true
}

// closesFence reports whether line closes a fence opened with char repeated length
// times: the same character, at least as long, and nothing else on the line.
func closesFence(line string, char byte, length int) bool {
	trimmed, indent := trimIndent(line)
	if indent > 3 {
		return false
	}
	count := 0
	for count < len(trimmed) && trimmed[count] == char {
		count++
	}
	if count < length {
		return false
	}
	return strings.TrimSpace(trimmed[count:]) == ""
}

// opensComment reports whether line starts an HTML comment block and returns what
// follows the opening marker, so the caller can see a comment closed on one line.
func opensComment(line string) (string, bool) {
	trimmed, indent := trimIndent(line)
	if indent > 3 || !strings.HasPrefix(trimmed, "<!--") {
		return "", false
	}
	return trimmed[len("<!--"):], true
}

// opensRawHTML reports the tag when line starts a raw HTML block.
func opensRawHTML(line string) (string, bool) {
	trimmed, indent := trimIndent(line)
	if indent > 3 || !strings.HasPrefix(trimmed, "<") {
		return "", false
	}
	lower := strings.ToLower(trimmed)
	for _, tag := range rawHTMLTags {
		if !strings.HasPrefix(lower, "<"+tag) {
			continue
		}
		switch rest := lower[len(tag)+1:]; {
		case rest == "":
			return tag, true
		case rest[0] == ' ', rest[0] == '>', rest[0] == '\t', rest[0] == '\n':
			return tag, true
		}
	}
	return "", false
}

// isSetextUnderline reports whether line turns the paragraph above it into a heading.
func isSetextUnderline(line string) bool {
	trimmed, indent := trimIndent(line)
	if indent > 3 {
		return false
	}
	body := strings.TrimRight(trimmed, " \t\n")
	if body == "" {
		return false
	}
	char := body[0]
	if char != '=' && char != '-' {
		return false
	}
	return strings.Trim(body, string(char)) == ""
}

// lastLineStart is the offset where the last line of buf begins.
func lastLineStart(buf []byte) int {
	if len(buf) == 0 {
		return 0
	}
	return bytes.LastIndexByte(buf[:len(buf)-1], '\n') + 1
}
