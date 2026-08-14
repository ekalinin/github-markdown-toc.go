package adapters

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
// An existing backup is never overwritten. The timestamp has one-second granularity,
// so a second run within the same second would otherwise replace the pristine copy
// with an already-rewritten one.
func (b *FileBackupper) Backup(ctx context.Context, file string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	// Resolve a symlink to its target, so the backup lands next to the real document
	// rather than next to the link. A path that cannot be resolved (e.g. it does not
	// exist) is handled below exactly as before.
	if resolved, resolveErr := filepath.EvalSymlinks(file); resolveErr == nil {
		file = resolved
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
	dst, err := os.OpenFile(backup, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
	if err != nil {
		return "", err
	}
	_, writeErr := dst.Write(data)
	closeErr := dst.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		// The exclusive create already claimed the path. Leaving a partial copy there
		// would block a retry with EEXIST and hide the error that actually happened.
		return "", errors.Join(err, os.Remove(backup))
	}
	return backup, nil
}
