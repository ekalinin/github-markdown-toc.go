package adapters

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
				if len(got) != 1 || got[0] != want {
					t.Errorf("Got=%v, want=[%v]", got, want)
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

func TestHTMLConverterSplitsLargeDocuments(t *testing.T) {
	var mu sync.Mutex
	var bodies [][]byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		mu.Lock()
		bodies = append(bodies, body)
		n := len(bodies)
		mu.Unlock()
		if _, err := fmt.Fprintf(w, "<h1>chunk %d</h1>", n); err != nil {
			t.Error(err)
		}
	}))
	defer srv.Close()

	var doc strings.Builder
	for doc.Len() <= 3*maxChunkBytes {
		doc.WriteString("# Heading\n\n")
		doc.WriteString(strings.Repeat("word ", 200))
		doc.WriteString("\n\n")
	}
	file := filepath.Join(t.TempDir(), "big.md")
	if err := os.WriteFile(file, []byte(doc.String()), 0644); err != nil {
		t.Fatal(err)
	}

	converter := NewHTMLConverterWithClient("", srv.URL, srv.Client(), NewLogger(false))
	got, err := converter.Convert(context.Background(), file)
	if err != nil {
		t.Fatal(err)
	}

	if len(got) < 4 {
		t.Fatalf("got %d chunks, want at least 4", len(got))
	}
	mu.Lock()
	defer mu.Unlock()
	if len(bodies) != len(got) {
		t.Fatalf("got %d requests for %d chunks, want one each", len(bodies), len(got))
	}
	for i, body := range bodies {
		if len(body) > maxChunkBytes {
			t.Errorf("request %d carried %d bytes, want at most %d", i, len(body), maxChunkBytes)
		}
	}
	for i, html := range got {
		if want := fmt.Sprintf("<h1>chunk %d</h1>", i+1); html != want {
			t.Errorf("chunk %d is %q, want %q", i, html, want)
		}
	}
}
