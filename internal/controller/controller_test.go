package controller

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
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

func TestProcessFilesPreservesInputOrder(t *testing.T) {
	files := []string{"first.md", "second.md", "third.md"}
	started := make(chan string, len(files))
	completed := make(chan string, len(files))
	releases := map[string]chan struct{}{
		"first.md":  make(chan struct{}),
		"second.md": make(chan struct{}),
		"third.md":  make(chan struct{}),
	}
	uc := useCaseFunc(func(_ context.Context, path string) (entity.Toc, error) {
		started <- path
		<-releases[path]
		completed <- path
		return entity.Toc{path}, nil
	})
	ctl := newTestController(Config{}, uc)

	var stdout bytes.Buffer
	done := make(chan error, 1)
	go func() {
		done <- ctl.ProcessFiles(context.Background(), &stdout, files...)
	}()

	for range files {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("not all use cases started")
		}
	}
	for _, path := range []string{"third.md", "second.md", "first.md"} {
		close(releases[path])
		select {
		case got := <-completed:
			if got != path {
				t.Fatalf("got completed path %q, want %q", got, path)
			}
		case <-time.After(time.Second):
			t.Fatalf("use case %q did not complete", path)
		}
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("controller did not finish")
	}

	if got, want := stdout.String(), "first.md\n\nsecond.md\n\nthird.md\n\n"; got != want {
		t.Errorf("got output %q, want %q", got, want)
	}
}

func TestProcessFilesSerialAndParallelOutputsMatch(t *testing.T) {
	files := []string{"first.md", "second.md", "third.md"}
	uc := useCaseFunc(func(_ context.Context, path string) (entity.Toc, error) {
		return entity.Toc{path}, nil
	})
	run := func(serial bool) string {
		t.Helper()
		ctl := newTestController(Config{Serial: serial}, uc)
		var stdout bytes.Buffer
		if err := ctl.ProcessFiles(context.Background(), &stdout, files...); err != nil {
			t.Fatal(err)
		}
		return stdout.String()
	}

	serialOutput := run(true)
	parallelOutput := run(false)
	if serialOutput != parallelOutput {
		t.Errorf("serial output %q does not match parallel output %q", serialOutput, parallelOutput)
	}
}

func TestProcessFilesLimitsParallelism(t *testing.T) {
	files := make([]string, maxParallelFiles+4)
	for i := range files {
		files[i] = fmt.Sprintf("file-%d.md", i)
	}

	started := make(chan struct{}, len(files))
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseWorkers := func() {
		releaseOnce.Do(func() { close(release) })
	}
	t.Cleanup(releaseWorkers)

	var active atomic.Int32
	var peak atomic.Int32
	uc := useCaseFunc(func(ctx context.Context, _ string) (entity.Toc, error) {
		current := active.Add(1)
		defer active.Add(-1)
		for observed := peak.Load(); current > observed; observed = peak.Load() {
			if peak.CompareAndSwap(observed, current) {
				break
			}
		}
		started <- struct{}{}
		select {
		case <-release:
			return entity.Toc{}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	})
	ctl := newTestController(Config{}, uc)
	done := make(chan error, 1)
	go func() {
		done <- ctl.ProcessFiles(context.Background(), io.Discard, files...)
	}()

	for i := 0; i < maxParallelFiles; i++ {
		select {
		case <-started:
		case err := <-done:
			t.Fatalf("controller returned before workers were released: %v", err)
		case <-time.After(time.Second):
			t.Fatal("parallel workers did not start")
		}
	}
	select {
	case <-started:
		t.Fatalf("more than %d use cases started concurrently", maxParallelFiles)
	case <-time.After(50 * time.Millisecond):
	}

	releaseWorkers()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("controller did not finish after workers were released")
	}
	if got := peak.Load(); got != maxParallelFiles {
		t.Errorf("got peak parallelism %d, want %d", got, maxParallelFiles)
	}
}

func TestProcessFilesCompletesAllJobsAfterErrors(t *testing.T) {
	files := make([]string, maxParallelFiles*2)
	for i := range files {
		files[i] = fmt.Sprintf("file-%d.md", i)
	}

	processingErr := errors.New("processing failed")
	var completed atomic.Int32
	uc := useCaseFunc(func(_ context.Context, path string) (entity.Toc, error) {
		completed.Add(1)
		if path == "file-0.md" {
			return nil, processingErr
		}
		return entity.Toc{path}, nil
	})
	ctl := newTestController(Config{}, uc)
	done := make(chan error, 1)
	go func() {
		done <- ctl.ProcessFiles(context.Background(), io.Discard, files...)
	}()

	select {
	case err := <-done:
		if !errors.Is(err, processingErr) {
			t.Fatalf("got error %v, want processing error", err)
		}
	case <-time.After(time.Second):
		t.Fatal("controller deadlocked after a use case error")
	}
	if got, want := completed.Load(), int32(len(files)); got != want {
		t.Errorf("completed %d jobs, want %d", got, want)
	}
}

func TestProcessFilesStopsWaitingAfterContextCancellation(t *testing.T) {
	files := make([]string, maxParallelFiles+4)
	for i := range files {
		files[i] = fmt.Sprintf("slow-%d.md", i)
	}

	started := make(chan struct{}, maxParallelFiles)
	uc := useCaseFunc(func(ctx context.Context, _ string) (entity.Toc, error) {
		started <- struct{}{}
		<-ctx.Done()
		return nil, ctx.Err()
	})
	ctl := newTestController(Config{}, uc)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)

	go func() {
		done <- ctl.ProcessFiles(ctx, io.Discard, files...)
	}()

	for i := 0; i < maxParallelFiles; i++ {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("parallel workers did not start")
		}
	}
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("got error %v, want context cancellation", err)
		}
		if !strings.Contains(err.Error(), "slow-0.md") {
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
