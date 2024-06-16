package adapters

import "testing"

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
			f, err := temper.CreateTemp("", "gh-toc-tests-*")
			if err != nil {
				t.Errorf("Got err=%v", err)
			}
			if !checker.Exists(f.Name()) {
				t.Errorf("File not exists, f=%v", f.Name())
			}
		})
	}
}
