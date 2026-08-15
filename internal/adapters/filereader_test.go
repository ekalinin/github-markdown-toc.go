package adapters

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileReaderRead(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(file, []byte("# Title\n"), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := NewFileReader().Read(context.Background(), file)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "# Title\n" {
		t.Errorf("got %q, want %q", got, "# Title\n")
	}
}

func TestFileReaderReadCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := NewFileReader().Read(ctx, "irrelevant.md"); err == nil {
		t.Fatal("got no error, want context cancellation")
	}
}
