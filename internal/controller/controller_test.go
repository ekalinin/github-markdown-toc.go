package controller

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
)

type useCaseFunc func(context.Context, string) (entity.Toc, error)

func (f useCaseFunc) Do(ctx context.Context, path string) (entity.Toc, error) {
	return f(ctx, path)
}

type loggerStub struct{}

func (loggerStub) Info(string, ...any) {}

func newTestController(cfg Config, uc useCase) *Controller {
	return New(cfg, uc, uc, uc, loggerStub{})
}

func TestProcessFilesPrintsSuccessfulResultsAndAggregatesErrors(t *testing.T) {
	firstErr := errors.New("first failure")
	secondErr := errors.New("second failure")
	uc := useCaseFunc(func(_ context.Context, path string) (entity.Toc, error) {
		switch path {
		case "good.md":
			return entity.Toc{"* [Good](#good)"}, nil
		case "first.md":
			return nil, firstErr
		default:
			return nil, secondErr
		}
	})
	ctl := newTestController(Config{Serial: true}, uc)

	var stdout bytes.Buffer
	err := ctl.ProcessFiles(context.Background(), &stdout, "good.md", "first.md", "second.md")
	if !errors.Is(err, firstErr) || !errors.Is(err, secondErr) {
		t.Fatalf("got error %v, want both document errors", err)
	}
	for _, path := range []string{"first.md", "second.md"} {
		if !strings.Contains(err.Error(), path) {
			t.Errorf("error %q does not contain path %q", err, path)
		}
	}
	if got, want := stdout.String(), "* [Good](#good)\n\n"; got != want {
		t.Errorf("got output %q, want %q", got, want)
	}
}

func TestProcessFilesAcceptsEmptyTOC(t *testing.T) {
	uc := useCaseFunc(func(context.Context, string) (entity.Toc, error) {
		return entity.Toc{}, nil
	})
	ctl := newTestController(Config{Serial: true}, uc)

	var stdout bytes.Buffer
	if err := ctl.ProcessFiles(context.Background(), &stdout, "empty.md"); err != nil {
		t.Fatal(err)
	}
	if got, want := stdout.String(), "\n"; got != want {
		t.Errorf("got output %q, want %q", got, want)
	}
}

func TestProcessFilesStopsWaitingAfterContextCancellation(t *testing.T) {
	started := make(chan struct{})
	uc := useCaseFunc(func(ctx context.Context, _ string) (entity.Toc, error) {
		close(started)
		<-ctx.Done()
		return nil, ctx.Err()
	})
	ctl := newTestController(Config{}, uc)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	go func() {
		done <- ctl.ProcessFiles(ctx, io.Discard, "slow.md")
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("use case did not start")
	}
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("got error %v, want context cancellation", err)
		}
		if !strings.Contains(err.Error(), "slow.md") {
			t.Errorf("error %q does not contain the document path", err)
		}
	case <-time.After(time.Second):
		t.Fatal("controller did not return after context cancellation")
	}
}

func TestProcessSTDINCleansUpTemporaryFile(t *testing.T) {
	processingErr := errors.New("processing failed")
	tests := []struct {
		name    string
		ucError error
	}{
		{name: "success"},
		{name: "processing error", ucError: processingErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tempPath string
			uc := useCaseFunc(func(_ context.Context, path string) (entity.Toc, error) {
				tempPath = path
				body, err := os.ReadFile(path)
				if err != nil {
					return nil, err
				}
				if got, want := string(body), "# Title"; got != want {
					return nil, fmt.Errorf("got stdin body %q, want %q", got, want)
				}
				return entity.Toc{}, tt.ucError
			})
			ctl := newTestController(Config{Serial: true}, uc)

			err := ctl.ProcessSTDIN(
				context.Background(),
				io.Discard,
				strings.NewReader("# Title"),
			)
			if tt.ucError == nil && err != nil {
				t.Fatal(err)
			}
			if tt.ucError != nil && !errors.Is(err, tt.ucError) {
				t.Fatalf("got error %v, want processing error", err)
			}
			if tempPath == "" {
				t.Fatal("use case did not receive a temporary file")
			}
			if _, statErr := os.Stat(tempPath); !errors.Is(statErr, os.ErrNotExist) {
				t.Errorf("stdin temporary file still exists: %v", statErr)
			}
		})
	}
}
