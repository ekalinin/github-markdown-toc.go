package skipheader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
)

type markdownProcessor interface {
	DoAs(context.Context, string, string) (entity.Toc, error)
}

type fileReader interface {
	Read(context.Context, string) ([]byte, error)
}

type fileTemper interface {
	CreateTemp(context.Context, string, string) (*os.File, error)
	Remove(string) error
}

type logger interface {
	Info(string, ...any)
}

// - cut everything up to and including <!--te-->
// - hand the trimmed copy to the inner use case, keeping the original display path
type SkipHeader struct {
	inner  markdownProcessor
	reader fileReader
	temper fileTemper
	log    logger
}

func New(inner markdownProcessor, reader fileReader, temper fileTemper, log logger) *SkipHeader {
	return &SkipHeader{inner: inner, reader: reader, temper: temper, log: log}
}

// trimToEndMarker drops every line up to and including the <!--te--> marker. The
// second result reports whether the marker was there at all.
func trimToEndMarker(content []byte) ([]byte, bool) {
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == entity.MarkerEnd {
			return []byte(strings.Join(lines[i+1:], "\n")), true
		}
	}
	return nil, false
}

func (uc *SkipHeader) Do(ctx context.Context, file string) (entity.Toc, error) {
	return uc.DoAs(ctx, file, file)
}

func (uc *SkipHeader) DoAs(ctx context.Context, file, displayPath string) (toc entity.Toc, err error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("skip header of %q: %w", file, err)
	}

	content, err := uc.reader.Read(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("read %q to skip its header: %w", file, err)
	}

	trimmed, found := trimToEndMarker(content)
	if !found {
		uc.log.Info("SkipHeader: no end marker, using the file as is", "file", file)
		return uc.inner.DoAs(ctx, file, displayPath)
	}

	tmpfile, err := uc.temper.CreateTemp(ctx, "", "ghtoc-skip-header-*.md")
	if err != nil {
		return nil, fmt.Errorf("create temporary file for %q: %w", file, err)
	}
	tempPath := tmpfile.Name()
	defer func() {
		if removeErr := uc.temper.Remove(tempPath); removeErr != nil {
			err = errors.Join(err, fmt.Errorf("remove temporary file for %q: %w", file, removeErr))
		}
	}()

	written, writeErr := tmpfile.Write(trimmed)
	closeErr := tmpfile.Close()
	switch {
	case writeErr != nil:
		return nil, fmt.Errorf("write temporary file for %q: %w", file, writeErr)
	case written != len(trimmed):
		return nil, fmt.Errorf("write temporary file for %q: %w", file, io.ErrShortWrite)
	case closeErr != nil:
		return nil, fmt.Errorf("close temporary file for %q: %w", file, closeErr)
	}

	uc.log.Info("SkipHeader: using the trimmed copy", "path", tempPath)
	return uc.inner.DoAs(ctx, tempPath, displayPath)
}
