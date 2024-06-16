package adapters

import "testing"

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
			if got := checker.Exists(tt.path); got != tt.want {
				t.Errorf("Got=%v, want=%v", got, tt.want)
			}
		})
	}
}
