package adapters

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileBackupperBackup(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "README.md")
	if err := os.WriteFile(file, []byte("original\n"), 0640); err != nil {
		t.Fatal(err)
	}
	stamp := time.Date(2026, 8, 12, 13, 45, 6, 0, time.UTC)

	got, err := NewFileBackupperX(func() time.Time { return stamp }).Backup(context.Background(), file)
	if err != nil {
		t.Fatal(err)
	}

	want := file + ".orig.2026-08-12_134506"
	if got != want {
		t.Errorf("got backup path %q, want %q", got, want)
	}
	data, err := os.ReadFile(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "original\n" {
		t.Errorf("got backup contents %q, want %q", data, "original\n")
	}
	info, err := os.Stat(got)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0640 {
		t.Errorf("got backup mode %v, want 0640", info.Mode().Perm())
	}
}

func TestFileBackupperMissingFile(t *testing.T) {
	dir := t.TempDir()

	if _, err := NewFileBackupper().Backup(context.Background(), filepath.Join(dir, "absent.md")); err == nil {
		t.Fatal("got no error, want a missing file error")
	}
}

func TestFileBackupperRefusesToOverwriteExistingBackup(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "README.md")
	if err := os.WriteFile(file, []byte("original\n"), 0644); err != nil {
		t.Fatal(err)
	}
	stamp := time.Date(2026, 8, 12, 13, 45, 6, 0, time.UTC)
	backupper := NewFileBackupperX(func() time.Time { return stamp })

	first, err := backupper.Backup(context.Background(), file)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("changed\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := backupper.Backup(context.Background(), file); err == nil {
		t.Fatal("got no error, want a refusal to overwrite the existing backup")
	}

	data, err := os.ReadFile(first)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "original\n" {
		t.Errorf("got backup contents %q, want the pristine original", data)
	}
}
