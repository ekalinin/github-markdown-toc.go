package adapters

import (
	"context"
	"os"
)

// FileReader reads a whole file into memory.
type FileReader struct{}

func NewFileReader() *FileReader {
	return &FileReader{}
}

func (f *FileReader) Read(ctx context.Context, file string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return os.ReadFile(file)
}
