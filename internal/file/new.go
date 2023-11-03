package file

import (
	ghtoc "github.com/ekalinin/github-markdown-toc.go"
	"github.com/ekalinin/github-markdown-toc.go/internal"
	"github.com/ekalinin/github-markdown-toc.go/internal/converter"
	"github.com/ekalinin/github-markdown-toc.go/internal/debugger"
	"github.com/ekalinin/github-markdown-toc.go/internal/grabber/jsongrabber"
	"github.com/ekalinin/github-markdown-toc.go/internal/grabber/regrabber"
)

// Tocer is an interface to get a TOC.
type Tocer interface {
	Toc() ghtoc.GHToc
}

// Option describes an option for a New function.
type Option func(*file)

// New creates new TOC grabber (Tocer) with path and multiple options.
func New(path string, opts ...Option) Tocer {
	f := file{
		Path:     path,
		Debugger: debugger.New(false, ""),
	}

	for _, opt := range opts {
		opt(&f)
	}

	switch t := detectType(path); t {
	case TypeLocalMD:
		f.SetPrefix("LocalMD: ")
		lmd := localMD(f)
		return &lmd
	case TypeRemoteMD:
		f.SetPrefix("RemoteMD: ")
		lmd := localMD(f)
		return &RemoteMD{
			LocalMD: lmd,
			HttpGet: internal.HttpGet,
		}
	case TypeRemoteHTML:
		f.SetPrefix("RemoteHTML: ")
		return &RemoteHTML{
			file:       f,
			HttpGet:    internal.HttpGetJson,
			tocGrabber: jsongrabber.New(f.Path),
		}
	}
	return nil
}

func localMD(f file) LocalMD {
	return LocalMD{
		file: f,
		converter2html: converter.NewMd2Html(
			f.GhToken, f.GhUrl,
			internal.HttpPost,
		),
		tocGrabber: regrabber.New(f.Path),
	}
}
