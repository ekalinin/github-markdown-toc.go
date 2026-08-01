package adapters

import (
	"context"
	"os"
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
