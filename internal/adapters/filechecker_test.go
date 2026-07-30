package adapters

import (
	"context"
	"testing"
)

func Test_FileChecker(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"FileChecker: exists", "./filechecker.go", true},
		{"FileChecker: not exists", "./filechecker_not_exists.go", false},
	}

	checker := NewFileCheck(NewLogger(false))
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			got, err := checker.Exists(context.Background(), tt.path)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("Got=%v, want=%v", got, tt.want)
			}
		})
	}
}
