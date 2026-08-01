package app

import (
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/controller"
	coretoc "github.com/ekalinin/github-markdown-toc.go/v2/internal/core/toc"
)

// copy of controller.Config
type Config struct {
	Files      []string
	Serial     bool
	HideHeader bool
	HideFooter bool
	StartDepth int
	Depth      int
	NoEscape   bool
	Indent     int
	Debug      bool
	GHToken    string
	GHUrl      string
	GHVersion  string
}

func (c Config) ToControllerConfig() controller.Config {
	return controller.Config{
		Files:      c.Files,
		Serial:     c.Serial,
		HideHeader: c.HideHeader,
		HideFooter: c.HideFooter,
		StartDepth: c.StartDepth,
		Depth:      c.Depth,
		NoEscape:   c.NoEscape,
		Indent:     c.Indent,
		Debug:      c.Debug,
		GHToken:    c.GHToken,
		GHUrl:      c.GHUrl,
		GHVersion:  c.GHVersion,
	}
}

func (c Config) ToRendererConfig() coretoc.Config {
	return coretoc.Config{
		AbsolutePaths: len(c.Files) > 0,
		StartDepth:    c.StartDepth,
		Depth:         c.Depth,
		Escape:        !c.NoEscape,
		Indent:        c.Indent,
	}
}
