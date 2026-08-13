package adapters

import (
	"context"
	"fmt"
	"os"
	"time"
)

// backupTimeLayout matches the suffix the bash gh-md-toc uses for its backups.
const backupTimeLayout = "2006-01-02_150405"

// FileBackupper copies a file next to itself before it gets rewritten.
type FileBackupper struct {
	now func() time.Time
}

func NewFileBackupper() *FileBackupper {
	return NewFileBackupperX(time.Now)
}

func NewFileBackupperX(now func() time.Time) *FileBackupper {
	return &FileBackupper{now: now}
}

// Backup copies file to "<file>.orig.<timestamp>" and returns the path of the copy.
func (b *FileBackupper) Backup(ctx context.Context, file string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	info, err := os.Stat(file)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}

	backup := fmt.Sprintf("%s.orig.%s", file, b.now().Format(backupTimeLayout))
	if err := os.WriteFile(backup, data, info.Mode().Perm()); err != nil {
		return "", err
	}
	return backup, nil
}
