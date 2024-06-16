package entity

import (
	"bytes"
	"testing"
)

func Test_TocPrint(t *testing.T) {
	tests := []struct {
		name string
		toc  *Toc
		want string
	}{
		{"Print", &Toc{"hello", "there"}, "hello\nthere\n\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer
			if err := tt.toc.Print(&b); err != nil {
				t.Errorf("failed print, err=%v", err)
			}
			if got := b.String(); got != tt.want {
				t.Errorf("Got=%s, want=%s", got, tt.want)
			}
		})
	}
}
