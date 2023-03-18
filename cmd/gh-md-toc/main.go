package main

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"

	ghtoc "github.com/ekalinin/github-markdown-toc.go"
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
	ghurl      = kingpin.Flag("github-url", "GitHub URL, default=https://api.github.com").String()
)

// check if there was an error (and panic if it was)
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Entry point
func main() {
	kingpin.Version(ghtoc.Version)
	kingpin.Parse()

	if *token == "" {
		*token = os.Getenv("GH_TOC_TOKEN")
	}

	if *ghurl == "" {
		*ghurl = os.Getenv("GH_TOC_URL")
	}

	pathsCount := len(*paths)

	// read file paths | urls from args
	absPathsInToc := pathsCount > 1
	ch := make(chan *ghtoc.GHToc, pathsCount)

	for _, p := range *paths {
		ghdoc := ghtoc.NewGHDoc(p, absPathsInToc, *startDepth, *depth, !*noEscape, *token, *indent, *debug)
		ghdoc.SetGHURL(*ghurl)

		if *serial {
			ch <- ghdoc.GetToc()
		} else {
			go func(path string) { ch <- ghdoc.GetToc() }(p)
		}
	}

	if !*hideHeader && pathsCount == 1 {
		fmt.Println()
		fmt.Println("Table of Contents")
		fmt.Println("=================")
		fmt.Println()
	}

	for i := 1; i <= pathsCount; i++ {
		toc := <-ch
		// #14, check if there's really TOC?
		if toc != nil {
			check(toc.Print(os.Stdout))
		}
	}

	// read md from stdin
	if pathsCount == 0 {
		bytes, err := io.ReadAll(os.Stdin)
		check(err)

		file, err := os.CreateTemp(os.TempDir(), "ghtoc")
		check(err)
		defer os.Remove(file.Name())

		check(os.WriteFile(file.Name(), bytes, 0644))
		check(ghtoc.NewGHDoc(file.Name(), false, *startDepth, *depth, !*noEscape, *token, *indent, *debug).
			SetGHURL(*ghurl).
			GetToc().
			Print(os.Stdout))
	}

	if !*hideFooter {
		fmt.Println("Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)")
	}
}
