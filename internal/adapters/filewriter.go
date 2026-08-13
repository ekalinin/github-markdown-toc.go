package adapters

import (
	"context"
	"os"
	"path/filepath"
)

type FileWriter struct{}

func NewFileWriter() *FileWriter {
	return &FileWriter{}
}

func (f *FileWriter) Write(ctx context.Context, file string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return os.WriteFile(file, data, 0644)
}

// WriteAtomic writes data through a temporary file in the same directory and renames
// it over the target, so a failed write can never truncate the original. The existing
// file mode is preserved.
func (f *FileWriter) WriteAtomic(ctx context.Context, file string, data []byte) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}

	perm := os.FileMode(0644)
	if info, statErr := os.Stat(file); statErr == nil {
		perm = info.Mode().Perm()
	}

	tmp, err := os.CreateTemp(filepath.Dir(file), filepath.Base(file)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		if err != nil {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err = tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = os.Chmod(tmpPath, perm); err != nil {
		return err
	}
	err = os.Rename(tmpPath, file)
	return err
}
