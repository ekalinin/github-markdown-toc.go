package insertmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
)

// createdBy is the attribution written into the document, next to the TOC.
const createdBy = "<!-- Created by https://github.com/ekalinin/github-markdown-toc.go -->"

// useCase is the local-document pipeline InsertMd wraps. It takes a display path
// because InsertMd needs the TOC rendered with bare anchors: the TOC is written into
// the document itself, and GitHub resolves relative links against that document's own
// directory, so a path prefix there would point somewhere else.
type useCase interface {
	DoAs(context.Context, string, string) (entity.Toc, error)
}

type fileReader interface {
	Read(context.Context, string) ([]byte, error)
}

type atomicWriter interface {
	WriteAtomic(context.Context, string, []byte) error
}

type fileBackupper interface {
	Backup(context.Context, string) (string, error)
}

type stamper interface {
	Stamp() string
}

type notifier interface {
	Notify(string, ...any)
}

type logger interface {
	Info(string, ...any)
}

// Config controls how the TOC block is written back into the document.
type Config struct {
	NoBackup   bool
	HideFooter bool
}

// - delegate to the inner use case to build the TOC
// - back up the original file
// - rewrite the block between <!--ts--> and <!--te-->
type InsertMd struct {
	cfg       Config
	inner     useCase
	reader    fileReader
	writer    atomicWriter
	backupper fileBackupper
	stamp     stamper
	notify    notifier
	log       logger
}

func New(cfg Config, inner useCase, reader fileReader, writer atomicWriter,
	backupper fileBackupper, stamp stamper, notify notifier, log logger) *InsertMd {
	return &InsertMd{
		cfg:       cfg,
		inner:     inner,
		reader:    reader,
		writer:    writer,
		backupper: backupper,
		stamp:     stamp,
		notify:    notify,
		log:       log,
	}
}

// block renders what goes between the markers: the TOC itself plus, unless the
// footer is hidden, the attribution and the signature.
func (uc *InsertMd) block(toc entity.Toc) []byte {
	lines := make([]string, 0, len(toc)+3)
	lines = append(lines, toc...)
	if !uc.cfg.HideFooter {
		lines = append(lines, "", createdBy, uc.stamp.Stamp())
	}
	return []byte(strings.Join(lines, "\n"))
}

func (uc *InsertMd) Do(ctx context.Context, file string) (entity.Toc, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("insert TOC into %q: %w", file, err)
	}

	uc.log.Info("InsertMD: start", "file", file)
	toc, err := uc.inner.DoAs(ctx, file, "")
	if err != nil {
		return nil, err
	}

	content, err := uc.reader.Read(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("read %q for TOC insertion: %w", file, err)
	}

	// Validate before the backup, so a document without markers leaves nothing behind.
	updated, err := replaceBetweenMarkers(content, uc.block(toc))
	if err != nil {
		return nil, fmt.Errorf("insert TOC into %q: %w", file, err)
	}

	if !uc.cfg.NoBackup {
		backup, backupErr := uc.backupper.Backup(ctx, file)
		if backupErr != nil {
			return nil, fmt.Errorf("back up %q: %w", file, backupErr)
		}
		uc.notify.Notify("!! Origin version of the file: %q", backup)
	}

	if err := uc.writer.WriteAtomic(ctx, file, updated); err != nil {
		return nil, fmt.Errorf("insert TOC into %q: %w", file, err)
	}
	uc.notify.Notify("!! TOC was added into: %q", file)

	uc.log.Info("InsertMD: done.")
	return toc, nil
}
