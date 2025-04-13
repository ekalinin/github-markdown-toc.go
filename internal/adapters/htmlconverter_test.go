package adapters

import (
	"errors"
	"testing"
)

type fakePoster struct {
	gotURL   string
	gotToken string
	gotPath  string
	retBody  string
	retErr   error
}

func (p *fakePoster) Post(url, token, path string) (string, error) {
	p.gotPath = path
	p.gotToken = token
	p.gotURL = url
	return p.retBody, p.retErr
}

func Test_HTMLConverter(t *testing.T) {

	token, url, path := "xx-token", "gh-url", "html-file"
	want := "html res"
	tests := []struct {
		name   string
		poster *fakePoster
		failed bool
	}{
		{"Convert ok", &fakePoster{retBody: want}, false},
		{"Convert fail", &fakePoster{retErr: errors.New("failed")}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewHTMLConverterX(token, url,
				tt.poster, NewLogger(false))

			got, err := converter.Convert(path)
			if tt.failed {
				if err == nil {
					t.Errorf("Should be failed, but no errors.")
				}
				if err.Error() != "failed" {
					t.Errorf("Error is not the same.")
				}
			}

			if !tt.failed {
				if got != want {
					t.Errorf("Got=%v, want=%v", got, want)
				}
				if got := tt.poster.gotPath; got != path {
					t.Errorf("Got=%v, want=%v", got, path)
				}
				if got := tt.poster.gotToken; got != token {
					t.Errorf("Got=%v, want=%v", got, token)
				}
				if got, want := tt.poster.gotURL, url+"/markdown/raw"; got != want {
					t.Errorf("Got=%v, want=%v", got, want)
				}
			}
		})
	}
}

func Test_HTMLConverterX(t *testing.T) {
	converter := NewHTMLConverter("gh-token", "gh-url", NewLogger(false))
	_, ok := converter.poster.(*RemotePoster)
	if !ok {
		t.Errorf("converter is not of type RemotePoster")
	}
}
