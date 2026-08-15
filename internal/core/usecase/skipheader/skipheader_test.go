package skipheader

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
)

type innerSpy struct {
	gotFile        string
	gotDisplayPath string
	gotContent     string
}

func (s *innerSpy) DoAs(_ context.Context, file, displayPath string) (entity.Toc, error) {
	s.gotFile = file
	s.gotDisplayPath = displayPath
	if data, err := os.ReadFile(file); err == nil {
		s.gotContent = string(data)
	}
	return entity.Toc{"* [Section](#section)"}, nil
}

type readerStub struct{ data []byte }

func (s readerStub) Read(context.Context, string) ([]byte, error) { return s.data, nil }

// temperStub keeps this core test free of the adapters package.
type temperStub struct{}

func (temperStub) CreateTemp(_ context.Context, dir, pattern string) (*os.File, error) {
	return os.CreateTemp(dir, pattern)
}

func (temperStub) Remove(path string) error { return os.Remove(path) }

type loggerStub struct{}

func (loggerStub) Info(string, ...any) {}

func TestTrimToEndMarker(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		found   bool
	}{
		{
			name:    "drops the existing TOC block",
			content: "# Title\n<!--ts-->\n* [Title](#title)\n<!--te-->\n\n## Section\n",
			want:    "\n## Section\n",
			found:   true,
		},
		{
			name:    "tolerates indentation",
			content: "  <!--te-->\n## Section\n",
			want:    "## Section\n",
			found:   true,
		},
		{
			name:    "no marker",
			content: "# Title\n## Section\n",
			found:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := trimToEndMarker([]byte(tt.content))
			if found != tt.found {
				t.Fatalf("got found=%v, want %v", found, tt.found)
			}
			if found && string(got) != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSkipHeaderPassesTrimmedCopyDown(t *testing.T) {
	content := "# Title\n<!--ts-->\n* [Title](#title)\n<!--te-->\n\n## Section\n"
	inner := &innerSpy{}
	uc := New(inner, readerStub{data: []byte(content)}, temperStub{}, loggerStub{})

	if _, err := uc.Do(context.Background(), "README.md"); err != nil {
		t.Fatal(err)
	}
	if inner.gotContent != "\n## Section\n" {
		t.Errorf("got inner content %q, want the trimmed document", inner.gotContent)
	}
	if inner.gotDisplayPath != "README.md" {
		t.Errorf("got display path %q, want the original file", inner.gotDisplayPath)
	}
	if _, err := os.Stat(inner.gotFile); !os.IsNotExist(err) {
		t.Errorf("got temporary file %q still on disk, want it removed", inner.gotFile)
	}
}

func TestSkipHeaderWithoutMarkerUsesTheFileAsIs(t *testing.T) {
	inner := &innerSpy{}
	uc := New(inner, readerStub{data: []byte("# Title\n")}, temperStub{}, loggerStub{})

	if _, err := uc.DoAs(context.Background(), "README.md", "https://example.com/README.md"); err != nil {
		t.Fatal(err)
	}
	if inner.gotFile != "README.md" {
		t.Errorf("got inner file %q, want the original path", inner.gotFile)
	}
	if inner.gotDisplayPath != "https://example.com/README.md" {
		t.Errorf("got display path %q, want it forwarded unchanged", inner.gotDisplayPath)
	}
}

// failingRemoveTemper hands out real temp files but cannot remove them.
type failingRemoveTemper struct{ err error }

func (t failingRemoveTemper) CreateTemp(_ context.Context, dir, pattern string) (*os.File, error) {
	return os.CreateTemp(dir, pattern)
}

func (t failingRemoveTemper) Remove(string) error { return t.err }

func TestSkipHeaderReportsCleanupFailure(t *testing.T) {
	removeErr := errors.New("remove failed")
	inner := &innerSpy{}
	uc := New(inner, readerStub{data: []byte("# Title\n<!--te-->\n\n## Section\n")},
		failingRemoveTemper{err: removeErr}, loggerStub{})

	_, err := uc.Do(context.Background(), "README.md")
	if inner.gotFile != "" {
		t.Cleanup(func() { _ = os.Remove(inner.gotFile) })
	}
	if !errors.Is(err, removeErr) {
		t.Fatalf("got error %v, want it to carry the removal failure", err)
	}
	if inner.gotContent != "\n## Section\n" {
		t.Errorf("got inner content %q, want the trimmed document - the inner call still ran", inner.gotContent)
	}
}

// closedFileTemper returns a temp file that is already closed, so writing to it fails.
type closedFileTemper struct{ removed *bool }

func (t closedFileTemper) CreateTemp(_ context.Context, dir, pattern string) (*os.File, error) {
	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return nil, err
	}
	if err := file.Close(); err != nil {
		return nil, err
	}
	return file, nil
}

func (t closedFileTemper) Remove(path string) error {
	*t.removed = true
	return os.Remove(path)
}

func TestSkipHeaderRemovesTempFileWhenTheWriteFails(t *testing.T) {
	removed := false
	inner := &innerSpy{}
	uc := New(inner, readerStub{data: []byte("# Title\n<!--te-->\n## Section\n")},
		closedFileTemper{removed: &removed}, loggerStub{})

	if _, err := uc.Do(context.Background(), "README.md"); err == nil {
		t.Fatal("got no error, want the write to an already-closed file to fail")
	}
	if !removed {
		t.Error("got no removal, want the temp file cleaned up after the write failed")
	}
	if inner.gotFile != "" {
		t.Errorf("got the inner use case called with %q, want it never reached", inner.gotFile)
	}
}
