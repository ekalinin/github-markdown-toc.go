package app

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"testing"

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
