package controller

import (
	"context"
	"fmt"
	"io"
	"os"
)

func (ctl *Controller) ProcessSTDIN(ctx context.Context, stdout io.Writer, stdin io.Reader) error {
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
	defer func() {
		if err := os.Remove(file.Name()); err != nil {
			_, _ = fmt.Fprintln(stdout, "Error during file delete:", err)
		}
	}()

	err = os.WriteFile(file.Name(), bytes, 0644)
	if err != nil {
		return fmt.Errorf("write standard input temporary file %q: %w", file.Name(), err)
	}

	if err := ctl.ProcessFiles(ctx, stdout, file.Name()); err != nil {
		return fmt.Errorf("process standard input: %w", err)
	}
	return nil
}
