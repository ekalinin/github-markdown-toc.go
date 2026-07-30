package remotemd

import (
	"context"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase/config"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase/localmd"
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
}

func (s temperStub) CreateTemp(ctx context.Context, dir, pattern string) (*os.File, error) {
	return s.create(ctx, dir, pattern)
}

type writerStub struct {
	err error
}

func (s writerStub) Write(context.Context, string, []byte) error {
	return s.err
}

type checkerStub struct {
	exists bool
}

func (s checkerStub) Exists(context.Context, string) (bool, error) {
	return s.exists, nil
}

type converterStub struct {
	err error
}

func (s converterStub) Convert(context.Context, string) (string, error) {
	return "<h1>Title</h1>", s.err
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

func newUseCase(
	t *testing.T,
	getter getterStub,
	temper temperStub,
	writer writerStub,
	converter converterStub,
	grabber grabberStub,
) *RemoteMd {
	t.Helper()
	local := localmd.New(
		config.Config{},
		checkerStub{exists: true},
		writer,
		converter,
		grabber,
		loggerStub{},
	)
	return New(config.Config{}, getter, local, temper, writer, loggerStub{})
}

func TestDoReturnsTOC(t *testing.T) {
	want := entity.Toc{"* [Title](#title)"}
	uc := newUseCase(
		t,
		getterStub{body: []byte("# Title"), contentType: "text/plain"},
		createTemper(t),
		writerStub{},
		converterStub{},
		grabberStub{toc: &want},
	)

	got, err := uc.Do(context.Background(), "https://example.com/README.md")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("got TOC %v, want %v", got, want)
	}
}

func TestDoPropagatesDownloadError(t *testing.T) {
	dependencyErr := errors.New("download failed")
	uc := newUseCase(
		t,
		getterStub{err: dependencyErr},
		createTemper(t),
		writerStub{},
		converterStub{},
		grabberStub{},
	)

	const documentURL = "https://example.com/README.md"
	_, err := uc.Do(context.Background(), documentURL)
	if !errors.Is(err, dependencyErr) {
		t.Fatalf("got error %v, want download error", err)
	}
	if !strings.Contains(err.Error(), documentURL) {
		t.Errorf("error %q does not contain the document URL", err)
	}
}

func TestDoPropagatesTemporaryFileError(t *testing.T) {
	dependencyErr := errors.New("temporary file failed")
	uc := newUseCase(
		t,
		getterStub{body: []byte("# Title"), contentType: "text/plain"},
		temperStub{
			create: func(context.Context, string, string) (*os.File, error) {
				return nil, dependencyErr
			},
		},
		writerStub{},
		converterStub{},
		grabberStub{},
	)

	_, err := uc.Do(context.Background(), "https://example.com/README.md")
	if !errors.Is(err, dependencyErr) {
		t.Fatalf("got error %v, want temporary file error", err)
	}
}

func TestDoPropagatesWriterError(t *testing.T) {
	dependencyErr := errors.New("write failed")
	uc := newUseCase(
		t,
		getterStub{body: []byte("# Title"), contentType: "text/plain"},
		createTemper(t),
		writerStub{err: dependencyErr},
		converterStub{},
		grabberStub{},
	)

	_, err := uc.Do(context.Background(), "https://example.com/README.md")
	if !errors.Is(err, dependencyErr) {
		t.Fatalf("got error %v, want writer error", err)
	}
}

func TestDoKeepsRemotePathForLocalProcessingError(t *testing.T) {
	dependencyErr := errors.New("convert failed")
	uc := newUseCase(
		t,
		getterStub{body: []byte("# Title"), contentType: "text/plain"},
		createTemper(t),
		writerStub{},
		converterStub{err: dependencyErr},
		grabberStub{},
	)

	const documentURL = "https://example.com/README.md"
	_, err := uc.Do(context.Background(), documentURL)
	if !errors.Is(err, dependencyErr) {
		t.Fatalf("got error %v, want converter error", err)
	}
	if !strings.Contains(err.Error(), documentURL) {
		t.Errorf("error %q does not contain the original document URL", err)
	}
}

func TestDoRejectsUnexpectedContentType(t *testing.T) {
	uc := newUseCase(
		t,
		getterStub{body: []byte("<html></html>"), contentType: "text/html"},
		createTemper(t),
		writerStub{},
		converterStub{},
		grabberStub{},
	)

	_, err := uc.Do(context.Background(), "https://example.com/README.md")
	if err == nil {
		t.Fatal("expected an error for an unexpected content type")
	}
}
