package adapters

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func Test_FileWriter(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"Written"},
	}
	writer := NewFileWriter()
	checker := NewFileCheck(NewLogger(false))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "tmp-for-test")
			test_data := "some-test"
			ctx := context.Background()
			err := writer.Write(ctx, file, []byte(test_data))
			if err != nil {
				t.Errorf("Got err=%v", err)
			}
			exists, err := checker.Exists(ctx, file)
			if err != nil {
				t.Fatal(err)
			}
			if !exists {
				t.Errorf("File not exists, f=%v", file)
			}
			data, err := os.ReadFile(file)
			if err != nil {
				t.Errorf("Got read err=%v", err)
			}
			if got := string(data); got != test_data {
				t.Errorf("Got=%v, want=%v", got, test_data)
			}
			err = os.Remove(file)
			if err != nil {
				t.Errorf("Error on delete file=%v err=%v", file, err)
			}
		})
	}
}

func TestFileWriterWriteAtomicPreservesMode(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "README.md")
	if err := os.WriteFile(file, []byte("old\n"), 0640); err != nil {
		t.Fatal(err)
	}

	if err := NewFileWriter().WriteAtomic(context.Background(), file, []byte("new\n")); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new\n" {
		t.Errorf("got %q, want %q", data, "new\n")
	}
	info, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0640 {
		t.Errorf("got mode %v, want 0640", info.Mode().Perm())
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("got %d files in the directory, want only the target file", len(entries))
	}
}

func TestFileWriterWriteAtomicFailsWithoutDirectory(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "sub", "README.md")

	if err := NewFileWriter().WriteAtomic(context.Background(), file, []byte("new\n")); err == nil {
		t.Fatal("got no error, want a failure for a missing directory")
	}
	if _, err := os.Stat(dir + "/sub"); !os.IsNotExist(err) {
		t.Error("got a leftover directory, want nothing created")
	}
}

func TestFileWriterWriteAtomicRemovesTempFileOnFailure(t *testing.T) {
	dir := t.TempDir()
	// Renaming onto a directory fails, and by then the temp file exists.
	target := filepath.Join(dir, "target")
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatal(err)
	}

	if err := NewFileWriter().WriteAtomic(context.Background(), target, []byte("new\n")); err == nil {
		t.Fatal("got no error, want the rename onto a directory to fail")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "target" {
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		t.Errorf("got directory entries %v, want only the target - the temp file must be removed", names)
	}
}
