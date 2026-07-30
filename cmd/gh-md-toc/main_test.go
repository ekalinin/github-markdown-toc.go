package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunReturnsNonZeroForMissingFile(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing.md")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run(
		context.Background(),
		[]string{"--hide-header", missingPath},
		&stdout,
		&stderr,
	)

	if code == 0 {
		t.Fatal("got exit code 0, want a non-zero exit code")
	}
	if !strings.Contains(stderr.String(), missingPath) {
		t.Errorf("error output %q does not contain the document path", stderr.String())
	}
	if strings.Contains(stdout.String(), "Created by") {
		t.Errorf("footer was printed after an error: %q", stdout.String())
	}
}
