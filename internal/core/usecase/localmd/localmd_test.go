package localmd

import (
	"context"
	"errors"
	"io/fs"
	"slices"
	"strings"
	"testing"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
)

type checkerStub struct {
	exists bool
	err    error
}

func (s checkerStub) Exists(context.Context, string) (bool, error) {
	return s.exists, s.err
}

type writerStub struct {
	err error
}

func (s writerStub) Write(context.Context, string, []byte) error {
	return s.err
}

type converterStub struct {
	html string
	err  error
}

func (s converterStub) Convert(context.Context, string) (string, error) {
	return s.html, s.err
}

type grabberStub struct {
	toc *entity.Toc
	err error
}

func (s grabberStub) Grab(context.Context, string, string) (*entity.Toc, error) {
	return s.toc, s.err
}

type loggerStub struct{}

func (loggerStub) Info(string, ...any) {}

func TestDoReturnsTOC(t *testing.T) {
	want := entity.Toc{"* [Title](#title)"}
	uc := New(
		false,
		checkerStub{exists: true},
		writerStub{},
		converterStub{html: "<h1>Title</h1>"},
		grabberStub{toc: &want},
		loggerStub{},
	)

	got, err := uc.Do(context.Background(), "README.md")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("got TOC %v, want %v", got, want)
	}
}

func TestDoAcceptsEmptyTOC(t *testing.T) {
	empty := entity.Toc{}
	uc := New(
		false,
		checkerStub{exists: true},
		writerStub{},
		converterStub{},
		grabberStub{toc: &empty},
		loggerStub{},
	)

	got, err := uc.Do(context.Background(), "empty.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("got TOC %v, want an empty TOC", got)
	}
}

func TestDoPropagatesDependencyErrors(t *testing.T) {
	dependencyErr := errors.New("dependency failed")
	tests := []struct {
		name      string
		debug     bool
		checker   checkerStub
		writer    writerStub
		converter converterStub
		grabber   grabberStub
		wantText  string
	}{
		{
			name:     "checker",
			checker:  checkerStub{err: dependencyErr},
			wantText: "check local Markdown",
		},
		{
			name:      "converter",
			checker:   checkerStub{exists: true},
			converter: converterStub{err: dependencyErr},
			wantText:  "convert local Markdown",
		},
		{
			name:      "debug writer",
			debug:     true,
			checker:   checkerStub{exists: true},
			converter: converterStub{},
			writer:    writerStub{err: dependencyErr},
			wantText:  "write debug HTML",
		},
		{
			name:      "grabber",
			checker:   checkerStub{exists: true},
			converter: converterStub{},
			grabber:   grabberStub{err: dependencyErr},
			wantText:  "grab TOC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := New(tt.debug, tt.checker, tt.writer, tt.converter, tt.grabber, loggerStub{})
			_, err := uc.Do(context.Background(), "broken.md")
			if !errors.Is(err, dependencyErr) {
				t.Fatalf("got error %v, want dependency error", err)
			}
			if !strings.Contains(err.Error(), tt.wantText) {
				t.Errorf("error %q does not contain %q", err, tt.wantText)
			}
			if !strings.Contains(err.Error(), "broken.md") {
				t.Errorf("error %q does not contain the document path", err)
			}
		})
	}
}

func TestDoReturnsErrorForMissingFile(t *testing.T) {
	uc := New(
		false,
		checkerStub{},
		writerStub{},
		converterStub{},
		grabberStub{},
		loggerStub{},
	)

	_, err := uc.Do(context.Background(), "missing.md")
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("got error %v, want fs.ErrNotExist", err)
	}
	if !strings.Contains(err.Error(), "missing.md") {
		t.Errorf("error %q does not contain the document path", err)
	}
}

func TestDoReturnsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	uc := New(
		false,
		checkerStub{exists: true},
		writerStub{},
		converterStub{},
		grabberStub{},
		loggerStub{},
	)

	_, err := uc.Do(ctx, "README.md")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got error %v, want context cancellation", err)
	}
}
