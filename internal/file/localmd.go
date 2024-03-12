package file

import (
	"os"

	ghtoc "github.com/ekalinin/github-markdown-toc.go"
	"github.com/ekalinin/github-markdown-toc.go/internal/grabber"
)

var _ Tocer = (*LocalMD)(nil)

type Converter interface {
	Convert(path string) (string, error)
}

// LocalMD represents a local Markdown file.
// Main steps to get a TOC:
// - convert md file to HTML (via GH API)
// - grab TOC from HTML
type LocalMD struct {
	file
	converter2html Converter
	tocGrabber     grabber.Grabber
}

// Toc converts MD file into HTML & grab TOC.
func (md *LocalMD) GetToc() ghtoc.GHToc {
	md.Log("local file: %s", md.Path)
	if _, err := os.Stat(md.Path); os.IsNotExist(err) {
		md.Log("local file is not exists.")
		return nil
	}

	md.Log("converting to html ...")
	html, err := md.converter2html.Convert(md.Path)
	if err != nil {
		md.Log("Failed to convert MD into HTML: %s", err)
		return nil
	}

	if md.Debug {
		htmlFile := md.Path + ".debug.html"
		md.Log("writing html file: %s", htmlFile)
		if err := os.WriteFile(htmlFile, []byte(html), 0644); err != nil {
			md.Log("writing html file error: %s", err)
			return nil
		}
	}

	md.Log("grabbing the TOC ...")
	toc, err := md.tocGrabber.Grab(html)
	if err != nil {
		md.Log("failed to grab TOC: %s", err)
		return nil
	}

	md.Log("done.")
	return toc
}
