package entity

import "testing"

func Test_GetType(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		result Type
	}{
		{"LocalMD", "./README.md", TypeLocalMD},
		{"RemoteMD", "https://raw.githubusercontent.com/ekalinin/github-markdown-toc.go/master/README.md", TypeRemoteMD},
		{"RemoteHTML", "https://github.com/ekalinin/github-markdown-toc.go/blob/master/README.md", TypeRemoteHTML},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetType(tt.path); got != tt.result {
				t.Errorf("Want=%d, got=%d", tt.result, got)
			}
		})
	}
}
