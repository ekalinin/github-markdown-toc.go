package adapters

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

type fakePoster struct {
	gotURL   string
	gotToken string
	gotPath  string
	gotBody  []byte
	retBody  string
	retErr   error
}

func (p *fakePoster) Post(_ context.Context, url, token, path string) (string, error) {
	p.gotPath = path
	p.gotToken = token
	p.gotURL = url
	return p.retBody, p.retErr
}

func (p *fakePoster) PostBody(_ context.Context, url, token string, body []byte) (string, error) {
	p.gotURL = url
	p.gotToken = token
	p.gotBody = body
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

			got, err := converter.Convert(context.Background(), path)
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

type posterStub struct{ err error }

func (s posterStub) Post(context.Context, string, string, string) (string, error) {
	return "", s.err
}

func (s posterStub) PostBody(context.Context, string, string, []byte) (string, error) {
	return "", s.err
}

func TestHTMLConverterRateLimitHint(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantHint bool
	}{
		{
			name:     "forbidden",
			err:      &HTTPStatusError{StatusCode: http.StatusForbidden, Body: "API rate limit exceeded"},
			wantHint: true,
		},
		{
			name:     "too many requests",
			err:      &HTTPStatusError{StatusCode: http.StatusTooManyRequests},
			wantHint: true,
		},
		{
			name:     "forbidden for a reason other than the rate limit",
			err:      &HTTPStatusError{StatusCode: http.StatusForbidden, Body: `{"message":"Bad credentials"}`},
			wantHint: false,
		},
		{
			name:     "server error keeps the bare message",
			err:      &HTTPStatusError{StatusCode: http.StatusInternalServerError},
			wantHint: false,
		},
		{
			name:     "transport error keeps the bare message",
			err:      errors.New("connection refused"),
			wantHint: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter := NewHTMLConverterX("", "https://api.github.com", posterStub{err: tt.err}, NewLogger(false))

			_, err := converter.Convert(context.Background(), "README.md")
			if !errors.Is(err, tt.err) {
				t.Fatalf("got error %v, want the original wrapped", err)
			}
			hasHint := strings.Contains(err.Error(), "GH_TOC_TOKEN")
			if hasHint != tt.wantHint {
				t.Errorf("got hint=%v in %q, want %v", hasHint, err, tt.wantHint)
			}
		})
	}
}
