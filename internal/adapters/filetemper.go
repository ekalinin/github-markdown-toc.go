package adapters

import "os"

type FileTemper struct {
}

func NewFileTemper() *FileTemper {
	return &FileTemper{}
}

func (f *FileTemper) CreateTemp(dir, pattern string) (*os.File, error) {
	return os.CreateTemp(dir, pattern)
}
