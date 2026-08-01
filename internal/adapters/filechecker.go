package adapters

import (
	"context"
	"os"
)

type FileChecker struct {
	log logger
}

func NewFileCheck(log logger) *FileChecker {
	return &FileChecker{log: log}
}

func (ch *FileChecker) Exists(ctx context.Context, file string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	ch.log.Info("FileChecker.Exists: start", "file", file)
	_, err := os.Stat(file)
	if err == nil {
		ch.log.Info("FileChecker.Exists: done", "res", true)
		return true, nil
	}
	if !os.IsNotExist(err) {
		return false, err
	}

	res := false
	ch.log.Info("FileChecker.Exists: done", "res", res)
	return res, nil
}
