package usecase

import (
	"github.com/ekalinin/github-markdown-toc.go/internal/core/ports"
	"github.com/ekalinin/github-markdown-toc.go/internal/core/usecase/config"
	"github.com/ekalinin/github-markdown-toc.go/internal/core/usecase/localmd"
	"github.com/ekalinin/github-markdown-toc.go/internal/core/usecase/remotehtml"
	"github.com/ekalinin/github-markdown-toc.go/internal/core/usecase/remotemd"
)

func New(cfg config.Config, checker ports.FileChecker, writer ports.FileWriter,
	converter ports.HTMLConverter, grabber ports.TocGrabber, log ports.Logger) (*localmd.LocalMd, *remotemd.RemoteMd, *remotehtml.RemoteHTML) {
	ucLocalMD := localmd.New(cfg, checker, writer, converter, grabber, log)
	ucRemoteMD := remotemd.New(cfg)
	ucRemoteHTML := remotehtml.New(cfg)

	return ucLocalMD, ucRemoteMD, ucRemoteHTML
}
