package main

import (
	"errors"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/app"
	coretoc "github.com/ekalinin/github-markdown-toc.go/v2/internal/core/toc"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/version"
)

const (
	cliName          = "gh-md-toc"
	defaultGitHubURL = "https://api.github.com"
	stdinMarker      = "-"
)

type cliOptions struct {
	paths      *[]string
	serial     *bool
	hideHeader *bool
	hideFooter *bool
	startDepth *int
	depth      *int
	noEscape   *bool
	token      *string
	indent     *int
	debug      *bool
	githubURL  *string
	reVersion  *string
}

func newCLI() (*kingpin.Application, cliOptions) {
	parser := kingpin.New(cliName, "")
	parser.Version(version.Full())

	pathsDesc := "Local path or URL of the document to grab TOC. Read MD from stdin if not entered."
	options := cliOptions{
		paths:      parser.Arg("path", pathsDesc).Strings(),
		serial:     parser.Flag("serial", "Grab TOCs in the serial mode").Bool(),
		hideHeader: parser.Flag("hide-header", "Hide TOC header").Bool(),
		hideFooter: parser.Flag("hide-footer", "Hide TOC footer").Bool(),
		startDepth: parser.Flag("start-depth", "Start including from this level. Defaults to 0 (include all levels)").Default("0").Int(),
		depth:      parser.Flag("depth", "How many levels of headings to include. Defaults to 0 (all)").Default("0").Int(),
		noEscape:   parser.Flag("no-escape", "Do not escape chars in sections").Bool(),
		token:      parser.Flag("token", "GitHub personal token").Envar("GH_TOC_TOKEN").String(),
		indent:     parser.Flag("indent", "Indent space of generated list").Default("2").Int(),
		debug:      parser.Flag("debug", "Show debug info").Bool(),
		githubURL: parser.Flag(
			"github-url",
			"GitHub URL. Default: "+defaultGitHubURL,
		).Default(defaultGitHubURL).Envar("GH_TOC_URL").String(),
		reVersion: parser.Flag(
			"re-version",
			"RegExp version. Default: "+version.GH_2024_03,
		).Default(version.GH_2024_03).Enum(version.SupportedGHVersions()...),
	}

	return parser, options
}

// extractStdinMarker removes the "-" STDIN marker from the argument list. The flag
// parser would otherwise try to read it as a flag.
func extractStdinMarker(args []string) ([]string, bool) {
	found := false
	rest := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == stdinMarker {
			found = true
			continue
		}
		rest = append(rest, arg)
	}
	return rest, found
}

func parseConfig(args []string) (app.Config, error) {
	args, useStdin := extractStdinMarker(args)

	parser, options := newCLI()
	if _, err := parser.Parse(args); err != nil {
		return app.Config{}, err
	}

	files := *options.paths
	if useStdin {
		if len(files) > 0 {
			return app.Config{}, errors.New(`the "-" STDIN marker cannot be combined with other paths`)
		}
		files = nil
	}

	return app.Config{
		Files:  files,
		Serial: *options.serial,
		Presentation: app.PresentationConfig{
			HideHeader: *options.hideHeader,
			HideFooter: *options.hideFooter,
		},
		GitHub: app.GitHubConfig{
			GHToken:   *options.token,
			GHUrl:     *options.githubURL,
			GHVersion: *options.reVersion,
		},
		TOC: coretoc.Config{
			StartDepth: *options.startDepth,
			Depth:      *options.depth,
			Escape:     !*options.noEscape,
			Indent:     *options.indent,
		},
		Debug: *options.debug,
	}, nil
}
