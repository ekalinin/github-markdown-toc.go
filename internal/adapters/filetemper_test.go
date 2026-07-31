package adapters

import (
	"context"
	"errors"
	"os"
	"testing"
)

func Test_FileTemper(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"Created"},
	}
	temper := NewFileTemper()
	checker := NewFileCheck(NewLogger(false))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			f, err := temper.CreateTemp(ctx, t.TempDir(), "gh-toc-tests-*")
			if err != nil {
				t.Errorf("Got err=%v", err)
			}
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
			exists, err := checker.Exists(ctx, f.Name())
			if err != nil {
				t.Fatal(err)
			}
			if !exists {
				t.Errorf("File not exists, f=%v", f.Name())
			}
			if err := temper.Remove(f.Name()); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(f.Name()); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("temporary file still exists: %v", err)
			}
		})
	}
}
