package remotemd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
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

func (s converterStub) Convert(context.Context, string) ([]string, error) {
	return []string{"<h1>Title</h1>"}, s.err
}

type grabberStub struct {
	toc     *entity.Toc
	err     error
	gotPath string
}

func (s *grabberStub) Grab(_ context.Context, path string, _ ...string) (*entity.Toc, error) {
	s.gotPath = path
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

func createTrackingTemper(t *testing.T) (temperStub, *string) {
	t.Helper()
	var path string
	tempDir := t.TempDir()
	return temperStub{
		create: func(_ context.Context, _, pattern string) (*os.File, error) {
			file, err := os.CreateTemp(tempDir, pattern)
			if err == nil {
				path = file.Name()
			}
			return file, err
		},
		remove: os.Remove,
	}, &path
}

func newUseCase(
	t *testing.T,
	getter getterStub,
	temper temperStub,
	writer writerStub,
	converter converterStub,
	grabber *grabberStub,
) *RemoteMd {
	t.Helper()
	local := localmd.New(
		false,
		checkerStub{exists: true},
		writer,
		converter,
		grabber,
		loggerStub{},
	)
	return New(getter, local, temper, loggerStub{})
}

func TestDoReturnsTOC(t *testing.T) {
	want := entity.Toc{"* [Title](#title)"}
	temper, tempPath := createTrackingTemper(t)
	uc := newUseCase(
		t,
		getterStub{body: []byte("# Title"), contentType: "text/plain; charset=utf-8"},
		temper,
		writerStub{},
		converterStub{},
		&grabberStub{toc: &want},
	)

	got, err := uc.Do(context.Background(), "https://example.com/README.md")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("got TOC %v, want %v", got, want)
	}
	if _, err := os.Stat(*tempPath); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("temporary file still exists after success: %v", err)
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
		&grabberStub{},
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
		&grabberStub{},
	)

	_, err := uc.Do(context.Background(), "https://example.com/README.md")
	if !errors.Is(err, dependencyErr) {
		t.Fatalf("got error %v, want temporary file error", err)
	}
}

func TestDoCleansUpAfterTemporaryFileWriteError(t *testing.T) {
	tempPath := filepath.Join(t.TempDir(), "remote.md")
	if err := os.WriteFile(tempPath, nil, 0600); err != nil {
		t.Fatal(err)
	}
	temper := temperStub{
		create: func(context.Context, string, string) (*os.File, error) {
			return os.Open(tempPath)
		},
		remove: os.Remove,
	}
	uc := newUseCase(
		t,
		getterStub{body: []byte("# Title"), contentType: "text/plain"},
		temper,
		writerStub{},
		converterStub{},
		&grabberStub{},
	)

	_, err := uc.Do(context.Background(), "https://example.com/README.md")
	if err == nil {
		t.Fatal("expected a temporary file write error")
	}
	if _, statErr := os.Stat(tempPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("temporary file still exists after write error: %v", statErr)
	}
}

func TestDoKeepsRemotePathForLocalProcessingError(t *testing.T) {
	dependencyErr := errors.New("convert failed")
	temper, tempPath := createTrackingTemper(t)
	uc := newUseCase(
		t,
		getterStub{body: []byte("# Title"), contentType: "text/plain"},
		temper,
		writerStub{},
		converterStub{err: dependencyErr},
		&grabberStub{},
	)

	const documentURL = "https://example.com/README.md"
	_, err := uc.Do(context.Background(), documentURL)
	if !errors.Is(err, dependencyErr) {
		t.Fatalf("got error %v, want converter error", err)
	}
	if !strings.Contains(err.Error(), documentURL) {
		t.Errorf("error %q does not contain the original document URL", err)
	}
	if _, statErr := os.Stat(*tempPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("temporary file still exists after processing error: %v", statErr)
	}
}

func TestDoRejectsUnexpectedContentType(t *testing.T) {
	uc := newUseCase(
		t,
		getterStub{body: []byte("<html></html>"), contentType: "text/html"},
		createTemper(t),
		writerStub{},
		converterStub{},
		&grabberStub{},
	)

	_, err := uc.Do(context.Background(), "https://example.com/README.md")
	if err == nil {
		t.Fatal("expected an error for an unexpected content type")
	}
}

func TestDoRejectsMalformedContentType(t *testing.T) {
	uc := newUseCase(
		t,
		getterStub{body: []byte("# Title"), contentType: `text/plain; charset="`},
		createTemper(t),
		writerStub{},
		converterStub{},
		&grabberStub{},
	)

	_, err := uc.Do(context.Background(), "https://example.com/README.md")
	if err == nil {
		t.Fatal("expected an error for a malformed content type")
	}
}

func TestDoReturnsTemporaryFileRemovalError(t *testing.T) {
	removeErr := errors.New("remove failed")
	temper := createTemper(t)
	temper.remove = func(string) error {
		return removeErr
	}
	want := entity.Toc{"* [Title](#title)"}
	uc := newUseCase(
		t,
		getterStub{body: []byte("# Title"), contentType: "text/plain"},
		temper,
		writerStub{},
		converterStub{},
		&grabberStub{toc: &want},
	)

	_, err := uc.Do(context.Background(), "https://example.com/README.md")
	if !errors.Is(err, removeErr) {
		t.Fatalf("got error %v, want removal error", err)
	}
}

func TestRemoteMdPassesURLAsDisplayPath(t *testing.T) {
	want := entity.Toc{"* [Title](#title)"}
	grabber := &grabberStub{toc: &want}
	uc := newUseCase(
		t,
		getterStub{body: []byte("# Title"), contentType: "text/plain"},
		createTemper(t),
		writerStub{},
		converterStub{},
		grabber,
	)

	const documentURL = "https://example.com/README.md"
	if _, err := uc.Do(context.Background(), documentURL); err != nil {
		t.Fatal(err)
	}
	if grabber.gotPath != documentURL {
		t.Errorf("got grabber path %q, want %q", grabber.gotPath, documentURL)
	}
}
