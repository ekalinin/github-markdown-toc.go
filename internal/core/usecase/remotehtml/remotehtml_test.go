package remotehtml

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase/config"
)

type getterStub struct {
	body        []byte
	contentType string
	err         error
}

func (s getterStub) Get(context.Context, string) ([]byte, string, error) {
	return s.body, s.contentType, s.err
}

type temperStub struct {
	create func(context.Context, string, string) (*os.File, error)
	remove func(string) error
}

func (s temperStub) CreateTemp(ctx context.Context, dir, pattern string) (*os.File, error) {
	return s.create(ctx, dir, pattern)
}

func (s temperStub) Remove(path string) error {
	if s.remove != nil {
		return s.remove(path)
	}
	return os.Remove(path)
}

type grabberStub struct {
	toc *entity.Toc
	err error
}

func (s grabberStub) Grab(context.Context, string) (*entity.Toc, error) {
	return s.toc, s.err
}

type loggerStub struct{}

func (loggerStub) Info(string, ...any) {}

func createTemper(t *testing.T) temperStub {
	t.Helper()
	return temperStub{
		create: func(_ context.Context, _, pattern string) (*os.File, error) {
			return os.CreateTemp(t.TempDir(), pattern)
		},
		remove: os.Remove,
	}
}

func TestDoReturnsTOC(t *testing.T) {
	want := entity.Toc{"* [Title](#title)"}
	uc := New(
		config.Config{},
		getterStub{body: []byte(`{"payload":{}}`), contentType: "application/json; charset=utf-8"},
		createTemper(t),
		grabberStub{toc: &want},
		loggerStub{},
	)

	got, err := uc.Do(context.Background(), "https://github.com/example/repo/blob/main/README.md")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("got TOC %v, want %v", got, want)
	}
}

func TestDoPropagatesDownloadError(t *testing.T) {
	dependencyErr := errors.New("download failed")
	uc := New(
		config.Config{},
		getterStub{err: dependencyErr},
		createTemper(t),
		grabberStub{},
		loggerStub{},
	)

	const documentURL = "https://github.com/example/repo/blob/main/README.md"
	_, err := uc.Do(context.Background(), documentURL)
	if !errors.Is(err, dependencyErr) {
		t.Fatalf("got error %v, want download error", err)
	}
	if !strings.Contains(err.Error(), documentURL) {
		t.Errorf("error %q does not contain the document URL", err)
	}
}

func TestDoPropagatesGrabberError(t *testing.T) {
	dependencyErr := errors.New("grab failed")
	uc := New(
		config.Config{},
		getterStub{body: []byte(`{"payload":{}}`), contentType: "application/json"},
		createTemper(t),
		grabberStub{err: dependencyErr},
		loggerStub{},
	)

	_, err := uc.Do(context.Background(), "https://github.com/example/repo/blob/main/README.md")
	if !errors.Is(err, dependencyErr) {
		t.Fatalf("got error %v, want grabber error", err)
	}
}

func TestDoPropagatesDebugFileCreateError(t *testing.T) {
	dependencyErr := errors.New("debug file failed")
	uc := New(
		config.Config{Debug: true},
		getterStub{body: []byte(`{"payload":{}}`), contentType: "application/json"},
		temperStub{
			create: func(context.Context, string, string) (*os.File, error) {
				return nil, dependencyErr
			},
		},
		grabberStub{},
		loggerStub{},
	)

	_, err := uc.Do(context.Background(), "https://github.com/example/repo/blob/main/README.md")
	if !errors.Is(err, dependencyErr) {
		t.Fatalf("got error %v, want debug file error", err)
	}
}

func TestDoCleansUpAfterDebugFileWriteError(t *testing.T) {
	tempPath := filepath.Join(t.TempDir(), "debug.json")
	if err := os.WriteFile(tempPath, nil, 0600); err != nil {
		t.Fatal(err)
	}
	uc := New(
		config.Config{Debug: true},
		getterStub{body: []byte(`{"payload":{}}`), contentType: "application/json"},
		temperStub{
			create: func(context.Context, string, string) (*os.File, error) {
				return os.Open(tempPath)
			},
			remove: os.Remove,
		},
		grabberStub{},
		loggerStub{},
	)

	_, err := uc.Do(context.Background(), "https://github.com/example/repo/blob/main/README.md")
	if err == nil {
		t.Fatal("expected a debug file write error")
	}
	if _, statErr := os.Stat(tempPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("failed debug artifact still exists: %v", statErr)
	}
}

func TestDoWritesSingleDebugArtifact(t *testing.T) {
	tempDir := t.TempDir()
	body := []byte(`{"payload":{"blob":{}}}`)
	empty := entity.Toc{}
	uc := New(
		config.Config{Debug: true},
		getterStub{body: body, contentType: "application/json"},
		temperStub{
			create: func(_ context.Context, _, pattern string) (*os.File, error) {
				return os.CreateTemp(tempDir, pattern)
			},
			remove: os.Remove,
		},
		grabberStub{toc: &empty},
		loggerStub{},
	)

	if _, err := uc.Do(context.Background(), "https://github.com/example/repo/blob/main/README.md"); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d debug files, want 1", len(entries))
	}
	if !strings.HasSuffix(entries[0].Name(), ".debug.json") {
		t.Errorf("debug artifact %q does not have the expected suffix", entries[0].Name())
	}
	got, err := os.ReadFile(filepath.Join(tempDir, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(body) {
		t.Errorf("got debug content %q, want %q", got, body)
	}
}

func TestDoRejectsUnexpectedContentType(t *testing.T) {
	uc := New(
		config.Config{},
		getterStub{body: []byte(`{"payload":{}}`), contentType: "text/html"},
		createTemper(t),
		grabberStub{},
		loggerStub{},
	)

	_, err := uc.Do(context.Background(), "https://github.com/example/repo/blob/main/README.md")
	if err == nil {
		t.Fatal("expected an error for an unexpected content type")
	}
}

func TestDoRejectsMalformedContentType(t *testing.T) {
	uc := New(
		config.Config{},
		getterStub{body: []byte(`{"payload":{}}`), contentType: `application/json; charset="`},
		createTemper(t),
		grabberStub{},
		loggerStub{},
	)

	_, err := uc.Do(context.Background(), "https://github.com/example/repo/blob/main/README.md")
	if err == nil {
		t.Fatal("expected an error for a malformed content type")
	}
}

func TestDoReturnsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	uc := New(
		config.Config{},
		getterStub{},
		createTemper(t),
		grabberStub{},
		loggerStub{},
	)

	_, err := uc.Do(ctx, "https://github.com/example/repo/blob/main/README.md")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got error %v, want context cancellation", err)
	}
}
