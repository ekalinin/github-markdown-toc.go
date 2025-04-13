package controller

import (
	"github.com/ekalinin/github-markdown-toc.go/internal/core/usecase/config"
)

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

func (c Config) ToUseCaseConfig() config.Config {
	return config.Config{
		Serial:       c.Serial,
		HideHeader:   c.HideHeader,
		HideFooter:   c.HideFooter,
		StartDepth:   c.StartDepth,
		Depth:        c.Depth,
		NoEscape:     c.NoEscape,
		Indent:       c.Indent,
		Debug:        c.Debug,
		GHToken:      c.GHToken,
		GHUrl:        c.GHUrl,
		GHVersion:    c.GHVersion,
		AbsPathInToc: len(c.Files) > 1,
	}
}
