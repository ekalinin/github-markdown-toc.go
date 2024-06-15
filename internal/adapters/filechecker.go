package adapters

import (
	"os"

	"github.com/ekalinin/github-markdown-toc.go/internal/core/ports"
)

type FileChecker struct {
	log ports.Logger
}

func NewFileCheck(log ports.Logger) *FileChecker {
	return &FileChecker{log: log}
}

func (ch *FileChecker) Exists(file string) bool {
	ch.log.Info("FileCheker.Exists: start", "file", file)
	_, err := os.Stat(file)
	res := !os.IsNotExist(err)
	ch.log.Info("FileCheker.Exists: done", "res", res)
	return res
}
