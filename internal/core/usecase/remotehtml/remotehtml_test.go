package remotehtml

import (
	"context"
	"errors"
	"os"
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

type writerStub struct {
	err error
}

func (s writerStub) Write(context.Context, string, []byte) error {
	return s.err
}

type temperStub struct {
	create func(context.Context, string, string) (*os.File, error)
}

func (s temperStub) CreateTemp(ctx context.Context, dir, pattern string) (*os.File, error) {
	return s.create(ctx, dir, pattern)
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
	}
}

func TestDoReturnsTOC(t *testing.T) {
	want := entity.Toc{"* [Title](#title)"}
	uc := New(
		config.Config{},
		getterStub{body: []byte(`{"payload":{}}`), contentType: "application/json"},
		writerStub{},
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
		writerStub{},
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
		getterStub{body: []byte(`{"payload":{}}`)},
		writerStub{},
		createTemper(t),
		grabberStub{err: dependencyErr},
		loggerStub{},
	)

	_, err := uc.Do(context.Background(), "https://github.com/example/repo/blob/main/README.md")
	if !errors.Is(err, dependencyErr) {
		t.Fatalf("got error %v, want grabber error", err)
	}
}

func TestDoPropagatesDebugFileErrors(t *testing.T) {
	dependencyErr := errors.New("debug file failed")
	tests := []struct {
		name   string
		temper temperStub
		writer writerStub
	}{
		{
			name: "create",
			temper: temperStub{
				create: func(context.Context, string, string) (*os.File, error) {
					return nil, dependencyErr
				},
			},
		},
		{
			name:   "write",
			temper: createTemper(t),
			writer: writerStub{err: dependencyErr},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := New(
				config.Config{Debug: true},
				getterStub{body: []byte(`{"payload":{}}`)},
				tt.writer,
				tt.temper,
				grabberStub{},
				loggerStub{},
			)

			_, err := uc.Do(context.Background(), "https://github.com/example/repo/blob/main/README.md")
			if !errors.Is(err, dependencyErr) {
				t.Fatalf("got error %v, want debug file error", err)
			}
		})
	}
}

func TestDoReturnsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	uc := New(
		config.Config{},
		getterStub{},
		writerStub{},
		createTemper(t),
		grabberStub{},
		loggerStub{},
	)

	_, err := uc.Do(ctx, "https://github.com/example/repo/blob/main/README.md")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got error %v, want context cancellation", err)
	}
}
