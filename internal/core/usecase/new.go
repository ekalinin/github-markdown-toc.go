package usecase

import (
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/ports"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase/config"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase/localmd"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase/remotehtml"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/usecase/remotemd"
)

func New(cfg config.Config,
	checker ports.FileChecker,
	writer ports.FileWriter,
	converter ports.HTMLConverter,
	grabberRe ports.TocGrabber,
	grabberJSON ports.TocGrabber,
	getter ports.RemoteGetter,
	temper ports.FileTemper,
	log ports.Logger) (*localmd.LocalMd, *remotemd.RemoteMd, *remotehtml.RemoteHTML) {

	ucLocalMD := localmd.New(cfg, checker, writer, converter, grabberRe, log)
	ucRemoteMD := remotemd.New(cfg, getter, ucLocalMD, temper, log)
	ucRemoteHTML := remotehtml.New(cfg, getter, temper, grabberJSON, log)

	return ucLocalMD, ucRemoteMD, ucRemoteHTML
}
