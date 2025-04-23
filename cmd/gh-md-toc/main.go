package main

import (
	"log"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/app"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/version"
)

var (
	pathsDesc  = "Local path or URL of the document to grab TOC. Read MD from stdin if not entered."
	paths      = kingpin.Arg("path", pathsDesc).Strings()
	serial     = kingpin.Flag("serial", "Grab TOCs in the serial mode").Bool()
	hideHeader = kingpin.Flag("hide-header", "Hide TOC header").Bool()
	hideFooter = kingpin.Flag("hide-footer", "Hide TOC footer").Bool()
	startDepth = kingpin.Flag("start-depth", "Start including from this level. Defaults to 0 (include all levels)").Default("0").Int()
	depth      = kingpin.Flag("depth", "How many levels of headings to include. Defaults to 0 (all)").Default("0").Int()
	noEscape   = kingpin.Flag("no-escape", "Do not escape chars in sections").Bool()
	token      = kingpin.Flag("token", "GitHub personal token").String()
	indent     = kingpin.Flag("indent", "Indent space of generated list").Default("2").Int()
	debug      = kingpin.Flag("debug", "Show debug info").Bool()
	ghurl      = kingpin.Flag("github-url", "GitHub URL, default=https://api.github.com").Default("https://api.github.com").String()
	reVersion  = kingpin.Flag("re-version", "RegExp version, default=0").Default(version.GH_2024_03).String()
)

// Entry point
func main() {
	kingpin.Version(version.Version)
	kingpin.Parse()

	if *token == "" {
		*token = os.Getenv("GH_TOC_TOKEN")
	}

	if *ghurl == "" {
		*ghurl = os.Getenv("GH_TOC_URL")
	}

	cfg := app.Config{
		Files:      *paths,
		Serial:     *serial,
		HideHeader: *hideHeader,
		HideFooter: *hideFooter,
		StartDepth: *startDepth,
		Depth:      *depth,
		NoEscape:   *noEscape,
		Indent:     *indent,
		Debug:      *debug,
		GHToken:    *token,
		GHUrl:      *ghurl,
		GHVersion:  *reVersion,
	}

	if err := app.New(cfg).Run(os.Stdout); err != nil {
		log.Fatal(err)
	}
}
