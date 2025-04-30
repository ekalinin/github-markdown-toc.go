package adapters

import (
	"os"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/ports"
)

type FileWriter struct {
	log ports.Logger
}

func NewFileWriter(log ports.Logger) *FileWriter {
	return &FileWriter{log: log}
}

func (f *FileWriter) Write(file string, data []byte) error {
	return os.WriteFile(file, data, 0644)
}
