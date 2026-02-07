package adapters

import (
	"fmt"
	"io"
	"os"
	"time"
)

type FileBackupper struct{}

func NewFileBackupper() *FileBackupper {
	return &FileBackupper{}
}

func (fb *FileBackupper) CreateBackup(filepath string) (string, error) {
	timestamp := time.Now().Format("2006-01-02_150405")
	backupPath := fmt.Sprintf("%s.%s", filepath, timestamp)

	src, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = src.Close()
	}()

	dst, err := os.Create(backupPath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = dst.Close()
	}()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return backupPath, nil
}
