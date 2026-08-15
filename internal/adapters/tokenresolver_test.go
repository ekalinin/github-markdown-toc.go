package adapters

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTokenResolverResolve(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		write    bool
		want     string
	}{
		{name: "no file", write: false, want: ""},
		{name: "token with newline", write: true, contents: "abc123\n", want: "abc123"},
		{name: "token with spaces", write: true, contents: "  abc123  ", want: "abc123"},
		{name: "empty file", write: true, contents: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.write {
				path := filepath.Join(dir, tokenFileName)
				if err := os.WriteFile(path, []byte(tt.contents), 0600); err != nil {
					t.Fatal(err)
				}
			}
			resolver := NewTokenResolverX(func() (string, error) {
				return filepath.Join(dir, "gh-md-toc"), nil
			})

			got, err := resolver.Resolve()
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got token %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTokenResolverIgnoresUnknownExecutablePath(t *testing.T) {
	resolver := NewTokenResolverX(func() (string, error) {
		return "", os.ErrNotExist
	})

	got, err := resolver.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("got token %q, want an empty string", got)
	}
}
