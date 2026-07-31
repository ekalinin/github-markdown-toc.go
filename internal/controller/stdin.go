package controller

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
)

func (ctl *Controller) ProcessSTDIN(ctx context.Context, stdout io.Writer, stdin io.Reader) (err error) {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("read standard input: %w", err)
	}

	bytes, err := io.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("read standard input: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("read standard input: %w", err)
	}

	file, err := os.CreateTemp(os.TempDir(), "ghtoc")
	if err != nil {
		return fmt.Errorf("create standard input temporary file: %w", err)
	}
	path := file.Name()
	fileOpen := true
	defer func() {
		if fileOpen {
			if closeErr := file.Close(); closeErr != nil {
				err = errors.Join(err, fmt.Errorf("close standard input temporary file %q: %w", path, closeErr))
			}
		}
		if removeErr := os.Remove(path); removeErr != nil {
			err = errors.Join(err, fmt.Errorf("remove standard input temporary file %q: %w", path, removeErr))
		}
	}()

	written, err := file.Write(bytes)
	if err != nil {
		return fmt.Errorf("write standard input temporary file %q: %w", path, err)
	}
	if written != len(bytes) {
		return fmt.Errorf("write standard input temporary file %q: %w", path, io.ErrShortWrite)
	}
	closeErr := file.Close()
	fileOpen = false
	if closeErr != nil {
		return fmt.Errorf("close standard input temporary file %q: %w", path, closeErr)
	}

	if err := ctl.ProcessFiles(ctx, stdout, path); err != nil {
		return fmt.Errorf("process standard input: %w", err)
	}
	return nil
}
