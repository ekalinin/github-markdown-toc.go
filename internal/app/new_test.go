package app

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	coretoc "github.com/ekalinin/github-markdown-toc.go/v2/internal/core/toc"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/version"
)

func TestNewDoesNotLogGitHubToken(t *testing.T) {
	var output bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&output, nil)))
	t.Cleanup(func() {
		slog.SetDefault(originalLogger)
	})

	const token = "secret-token"
	_, err := New(Config{
		Debug: true,
		GitHub: GitHubConfig{
			GHToken:   token,
			GHUrl:     "https://api.github.com",
			GHVersion: version.GH_2024_03,
		},
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}

	logOutput := output.String()
	if strings.Contains(logOutput, token) {
		t.Errorf("debug log contains GitHub token: %s", logOutput)
	}
	if !strings.Contains(logOutput, "token-configured=true") {
		t.Errorf("debug log does not report token presence safely: %s", logOutput)
	}
}

func TestNewRejectsUnknownRegexpVersion(t *testing.T) {
	_, err := New(Config{GitHub: GitHubConfig{GHVersion: "unknown"}}, io.Discard)
	if err == nil {
		t.Fatal("expected an error for an unsupported regexp version")
	}

	want := "initialize regexp grabber: unsupported GitHub regexp version \"unknown\""
	if !strings.Contains(err.Error(), want) {
		t.Errorf("got error %q, want it to contain %q", err, want)
	}
}

func TestNewSkipHeaderTrimsTheDocumentSentToGitHub(t *testing.T) {
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		gotBody = string(body)
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<h2 class="heading-element">Section</h2>`))
	}))
	defer server.Close()

	dir := t.TempDir()
	file := filepath.Join(dir, "README.md")
	content := "# Old Title\n<!--ts-->\n* [Old Title](#old-title)\n<!--te-->\n\n## Section\n"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	application, err := New(Config{
		Files:      []string{file},
		SkipHeader: true,
		GitHub:     GitHubConfig{GHUrl: server.URL, GHVersion: version.GH_2024_03},
		TOC:        coretoc.DefaultConfig(),
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if err := application.Run(context.Background(), &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(gotBody, "Old Title") {
		t.Errorf("got request body %q, want everything up to <!--te--> dropped", gotBody)
	}
	if !strings.Contains(gotBody, "## Section") {
		t.Errorf("got request body %q, want the content after <!--te--> kept", gotBody)
	}
}

func TestNewWithoutSkipHeaderSendsTheWholeDocument(t *testing.T) {
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		gotBody = string(body)
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<h2 class="heading-element">Section</h2>`))
	}))
	defer server.Close()

	dir := t.TempDir()
	file := filepath.Join(dir, "README.md")
	content := "# Old Title\n<!--ts-->\n* [Old Title](#old-title)\n<!--te-->\n\n## Section\n"
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	application, err := New(Config{
		Files:  []string{file},
		GitHub: GitHubConfig{GHUrl: server.URL, GHVersion: version.GH_2024_03},
		TOC:    coretoc.DefaultConfig(),
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if err := application.Run(context.Background(), &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(gotBody, "Old Title") {
		t.Errorf("got request body %q, want the untouched document", gotBody)
	}
}
