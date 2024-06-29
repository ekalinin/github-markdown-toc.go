package adapters

import "testing"

func Test_RemotePoster(t *testing.T) {
	want := "post ok"
	tests := []struct {
		name         string
		remotePoster *RemotePoster
		fake         bool
	}{
		{"Fake", NewRemotePosterX(&fakePoster{retBody: want}), true},
		{"Real", NewRemotePoster(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fake {
				got, err := tt.remotePoster.Post("url", "token", "path")
				if err != nil {
					t.Errorf("got err=%v", err)
				}
				if got != want {
					t.Errorf("got=%v, want=%v", got, want)
				}
			} else {
				p, ok := tt.remotePoster.poster.(*realPoster)
				if !ok {
					t.Errorf("should be used realPoster, got=%v", p)
				}
			}
		})
	}
}
