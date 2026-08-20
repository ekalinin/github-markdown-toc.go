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
	toc            entity.Toc
	err            error
	gotDisplayPath *string
}

func (s innerStub) DoAs(_ context.Context, _, displayPath string) (entity.Toc, error) {
	if s.gotDisplayPath != nil {
		*s.gotDisplayPath = displayPath
	}
	return s.toc, s.err
}

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

func TestInsertMdReadFailureLeavesFileAlone(t *testing.T) {
	readErr := errors.New("read failed")
	writer := &writerSpy{}
	backupper := &backupperSpy{}
	uc := New(Config{}, innerStub{toc: entity.Toc{"* [A](#a)"}}, readerStub{err: readErr}, writer,
		backupper, stamperStub{}, &notifierSpy{}, loggerStub{})

	if _, err := uc.Do(context.Background(), "README.md"); !errors.Is(err, readErr) {
		t.Fatalf("got error %v, want the read error", err)
	}
	if writer.got != nil {
		t.Errorf("got a write of %q, want none", writer.got)
	}
	if backupper.calls != 0 {
		t.Errorf("got %d backup calls, want none", backupper.calls)
	}
}

func TestInsertMdBackupFailureSkipsTheWrite(t *testing.T) {
	backupErr := errors.New("backup failed")
	writer := &writerSpy{}
	notify := &notifierSpy{}
	uc := New(Config{}, innerStub{toc: entity.Toc{"* [A](#a)"}},
		readerStub{data: []byte("<!--ts-->\n<!--te-->\n")}, writer,
		&backupperSpy{err: backupErr}, stamperStub{}, notify, loggerStub{})

	if _, err := uc.Do(context.Background(), "README.md"); !errors.Is(err, backupErr) {
		t.Fatalf("got error %v, want the backup error", err)
	}
	if writer.got != nil {
		t.Errorf("got a write of %q, want none - a failed backup must never be followed by a rewrite", writer.got)
	}
	if len(notify.messages) != 0 {
		t.Errorf("got messages %v, want none - nothing succeeded", notify.messages)
	}
}

func TestInsertMdWriteFailurePropagates(t *testing.T) {
	writeErr := errors.New("write failed")
	writer := &writerSpy{err: writeErr}
	notify := &notifierSpy{}
	uc := New(Config{NoBackup: true}, innerStub{toc: entity.Toc{"* [A](#a)"}},
		readerStub{data: []byte("<!--ts-->\n<!--te-->\n")}, writer,
		&backupperSpy{}, stamperStub{}, notify, loggerStub{})

	if _, err := uc.Do(context.Background(), "README.md"); !errors.Is(err, writeErr) {
		t.Fatalf("got error %v, want the write error", err)
	}
	if len(notify.messages) != 0 {
		t.Errorf("got messages %v, want none - the insert notice must not claim a write that failed", notify.messages)
	}
}

func TestInsertMdAsksTheInnerUseCaseForBareAnchors(t *testing.T) {
	var gotDisplayPath string
	uc := New(Config{NoBackup: true},
		innerStub{toc: entity.Toc{"* [A](#a)"}, gotDisplayPath: &gotDisplayPath},
		readerStub{data: []byte("<!--ts-->\n<!--te-->\n")}, &writerSpy{},
		&backupperSpy{}, stamperStub{}, &notifierSpy{}, loggerStub{})

	if _, err := uc.Do(context.Background(), "README.md"); err != nil {
		t.Fatal(err)
	}
	if gotDisplayPath != "" {
		t.Errorf("got display path %q, want an empty one so the TOC links to the document itself", gotDisplayPath)
	}
}
