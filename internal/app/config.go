package app

import coretoc "github.com/ekalinin/github-markdown-toc.go/v2/internal/core/toc"

// Config contains application-level options and configuration for its dependencies.
type Config struct {
	Files        []string
	Serial       bool
	SkipHeader   bool
	Presentation PresentationConfig
	GitHub       GitHubConfig
	TOC          coretoc.Config
	Insert       InsertConfig
	Debug        bool
}

// PresentationConfig controls the CLI output surrounding the generated TOC.
type PresentationConfig struct {
	HideHeader bool
	HideFooter bool
}

// GitHubConfig contains GitHub API and HTML layout settings.
type GitHubConfig struct {
	GHToken   string
	GHUrl     string
	GHVersion string
}

// InsertConfig controls rewriting the TOC inside the source document.
type InsertConfig struct {
	Enabled  bool
	NoBackup bool
}
