package adapters

import (
	"context"
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
			f, err := temper.CreateTemp(ctx, "", "gh-toc-tests-*")
			if err != nil {
				t.Errorf("Got err=%v", err)
			}
			exists, err := checker.Exists(ctx, f.Name())
			if err != nil {
				t.Fatal(err)
			}
			if !exists {
				t.Errorf("File not exists, f=%v", f.Name())
			}
		})
	}
}
