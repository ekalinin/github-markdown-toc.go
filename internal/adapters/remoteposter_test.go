package adapters

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

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
				testToken := "token-for-test"
				fileName, err := NewFileTemper().CreateTemp("", "example.*.txt")
				if err != nil {
					t.Error("Tmp file creation err=", err)
				}
				defer func() {
					if err := os.Remove(fileName.Name()); err != nil {
						t.Error("Tmp file deletion err=", err)
					}
				}()

				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != "POST" {
						t.Error("Should be POST")
					}
					tokenGot := r.Header.Get("Authorization")
					tokenWant := "token " + testToken
					if tokenGot != tokenWant {
						t.Error("Auth fail. Want token=", tokenWant, ", got=", tokenGot)
					}

					ctGot := r.Header.Get("Content-Type")
					ctWant := "text/plain;charset=utf-8"
					if ctGot != ctWant {
						t.Error("Content type fail. Want=", ctWant, ", but got=", ctGot)
					}
				}))
				defer srv.Close()

				if _, err := tt.remotePoster.Post(srv.URL, testToken, fileName.Name()); err != nil {
					t.Error("Should not be err, but got=", err)
				}
			}
		})
	}
}
