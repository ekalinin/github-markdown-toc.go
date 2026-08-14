package insertmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
)

type innerStub struct {
	toc entity.Toc
	err error
}

func (s innerStub) Do(context.Context, string) (entity.Toc, error) { return s.toc, s.err }

type readerStub struct {
	data []byte
	err  error
}

func (s readerStub) Read(context.Context, string) ([]byte, error) { return s.data, s.err }

type writerSpy struct {
	got []byte
	err error
}

func (s *writerSpy) WriteAtomic(_ context.Context, _ string, data []byte) error {
	s.got = data
	return s.err
}

type backupperSpy struct {
	calls int
	err   error
}

func (s *backupperSpy) Backup(_ context.Context, file string) (string, error) {
	s.calls++
	return file + ".orig.2026-08-12_134506", s.err
}

type stamperStub struct{}

func (stamperStub) Stamp() string { return "<!-- Added by: tester, at: 2026-08-12T13:45:06Z -->" }

type notifierSpy struct{ messages []string }

func (s *notifierSpy) Notify(format string, args ...any) {
	s.messages = append(s.messages, fmt.Sprintf(format, args...))
}

type loggerStub struct{}

func (loggerStub) Info(string, ...any) {}

func newTestInsertMd(cfg Config, content string, toc entity.Toc) (*InsertMd, *writerSpy, *backupperSpy, *notifierSpy) {
	writer := &writerSpy{}
	backupper := &backupperSpy{}
	notify := &notifierSpy{}
	uc := New(cfg, innerStub{toc: toc}, readerStub{data: []byte(content)}, writer,
		backupper, stamperStub{}, notify, loggerStub{})
	return uc, writer, backupper, notify
}

func TestInsertMdWritesBlockWithFooter(t *testing.T) {
	content := "# Title\n\n<!--ts-->\nstale\n<!--te-->\n\n## Section\n"
	uc, writer, backupper, notify := newTestInsertMd(Config{}, content, entity.Toc{"* [Title](#title)"})

	got, err := uc.Do(context.Background(), "README.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "* [Title](#title)" {
		t.Errorf("got TOC %v, want it returned unchanged", got)
	}

	want := "# Title\n\n<!--ts-->\n" +
		"* [Title](#title)\n" +
		"\n" +
		"<!-- Created by https://github.com/ekalinin/github-markdown-toc.go -->\n" +
		"<!-- Added by: tester, at: 2026-08-12T13:45:06Z -->\n" +
		"<!--te-->\n\n## Section\n"
	if string(writer.got) != want {
		t.Errorf("got written file\n%q\nwant\n%q", writer.got, want)
	}
	if backupper.calls != 1 {
		t.Errorf("got %d backup calls, want 1", backupper.calls)
	}
	if len(notify.messages) != 2 {
		t.Errorf("got messages %v, want the backup and the insert notice", notify.messages)
	}
}

func TestInsertMdHideFooter(t *testing.T) {
	content := "<!--ts-->\n<!--te-->\n"
	uc, writer, _, _ := newTestInsertMd(Config{HideFooter: true}, content, entity.Toc{"* [Title](#title)"})

	if _, err := uc.Do(context.Background(), "README.md"); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(writer.got), "Added by") {
		t.Errorf("got written file %q, want no signature comment", writer.got)
	}
}

func TestInsertMdNoBackup(t *testing.T) {
	content := "<!--ts-->\n<!--te-->\n"
	uc, _, backupper, notify := newTestInsertMd(Config{NoBackup: true}, content, entity.Toc{"* [A](#a)"})

	if _, err := uc.Do(context.Background(), "README.md"); err != nil {
		t.Fatal(err)
	}
	if backupper.calls != 0 {
		t.Errorf("got %d backup calls, want none", backupper.calls)
	}
	if len(notify.messages) != 1 {
		t.Errorf("got messages %v, want only the insert notice", notify.messages)
	}
}

func TestInsertMdMissingMarkersLeavesFileAlone(t *testing.T) {
	uc, writer, backupper, _ := newTestInsertMd(Config{}, "# Title\n", entity.Toc{"* [Title](#title)"})

	_, err := uc.Do(context.Background(), "README.md")
	if !errors.Is(err, ErrMarkersNotFound) {
		t.Fatalf("got error %v, want ErrMarkersNotFound", err)
	}
	if writer.got != nil {
		t.Errorf("got a write of %q, want none", writer.got)
	}
	if backupper.calls != 0 {
		t.Errorf("got %d backup calls, want none - validation runs first", backupper.calls)
	}
}

func TestInsertMdPropagatesInnerError(t *testing.T) {
	innerErr := errors.New("grab failed")
	uc := New(Config{}, innerStub{err: innerErr}, readerStub{}, &writerSpy{},
		&backupperSpy{}, stamperStub{}, &notifierSpy{}, loggerStub{})

	if _, err := uc.Do(context.Background(), "README.md"); !errors.Is(err, innerErr) {
		t.Fatalf("got error %v, want the inner error", err)
	}
}
