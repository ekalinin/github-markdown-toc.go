package adapters

import (
	"context"
	"os"
)

type FileTemper struct {
}

func NewFileTemper() *FileTemper {
	return &FileTemper{}
}

func (f *FileTemper) CreateTemp(ctx context.Context, dir, pattern string) (*os.File, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return os.CreateTemp(dir, pattern)
}

func (f *FileTemper) Remove(path string) error {
	return os.Remove(path)
}
