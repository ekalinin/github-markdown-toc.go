package mdsplit

import (
	"bytes"
	"strings"
	"testing"
)

func TestSplitKeepsSmallDocumentWhole(t *testing.T) {
	content := []byte("# A\n\n# B\n")

	got, err := Split(content, 1024)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d chunks, want 1", len(got))
	}
	if !bytes.Equal(got[0], content) {
		t.Errorf("got chunk %q, want %q", got[0], content)
	}
}

func TestSplitCutsAtBlankLines(t *testing.T) {
	got, err := Split([]byte("# A\n\n# B\n"), 6)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"# A\n\n", "# B\n"}
	if len(got) != len(want) {
		t.Fatalf("got %d chunks (%q), want %d", len(got), got, len(want))
	}
	for i := range want {
		if string(got[i]) != want[i] {
			t.Errorf("chunk %d is %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSplitClosesAndReopensAFence(t *testing.T) {
	got, err := Split([]byte("```\n\na\n\nb\n```\n\n# B\n"), 12)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"```\n\na\n\n```\n", "```\nb\n```\n\n", "# B\n"}
	if len(got) != len(want) {
		t.Fatalf("got %d chunks (%q), want %d", len(got), got, len(want))
	}
	for i := range want {
		if string(got[i]) != want[i] {
			t.Errorf("chunk %d is %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSplitKeepsSetextHeadingWhole(t *testing.T) {
	got, err := Split([]byte("x\nText\n----\n"), 10)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"x\n", "Text\n----\n"}
	if len(got) != len(want) {
		t.Fatalf("got %d chunks (%q), want %d", len(got), got, len(want))
	}
	for i := range want {
		if string(got[i]) != want[i] {
			t.Errorf("chunk %d is %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSplitRejectsOversizedLine(t *testing.T) {
	_, err := Split([]byte(strings.Repeat("a", 100)+"\n"), 10)
	if err == nil {
		t.Fatal("got no error, want an oversized line error")
	}
	if !strings.Contains(err.Error(), "does not fit") {
		t.Errorf("error %q does not explain that the line is too long", err)
	}
}

// TestSplitKeepsBlocksBalanced is the invariant that matters: a chunk must never
// leave a fence or an HTML comment open, because content that was code or a comment
// in the source would then be parsed as Markdown and produce phantom headings.
func TestSplitKeepsBlocksBalanced(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"plain text", "# Heading\n\nsome words here\n\n"},
		{"fenced code with hashes", "```\n# not a heading\nmore code\n```\n\n"},
		{"html comment with hashes", "<!--\n# not a heading\nmore text\n-->\n\n"},
		{"pre block with hashes", "<pre>\n# not a heading\nmore text\n</pre>\n\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const maxBytes = 64
			doc := strings.Repeat(tt.body, 40)

			chunks, err := Split([]byte(doc), maxBytes)
			if err != nil {
				t.Fatal(err)
			}
			if len(chunks) < 2 {
				t.Fatalf("got %d chunks, want the document to be split", len(chunks))
			}
			for i, chunk := range chunks {
				if len(chunk) > maxBytes {
					t.Errorf("chunk %d is %d bytes, want at most %d", i, len(chunk), maxBytes)
				}
				text := string(chunk)
				if n := strings.Count(text, "```"); n%2 != 0 {
					t.Errorf("chunk %d leaves a fence open: %q", i, text)
				}
				if strings.Count(text, "<!--") != strings.Count(text, "-->") {
					t.Errorf("chunk %d leaves an HTML comment open: %q", i, text)
				}
				if strings.Count(text, "<pre>") != strings.Count(text, "</pre>") {
					t.Errorf("chunk %d leaves a pre block open: %q", i, text)
				}
			}
		})
	}
}
