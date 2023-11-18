package file

import (
	"os"
	"strings"

	ghtoc "github.com/ekalinin/github-markdown-toc.go"
)

var _ Tocer = (*RemoteMD)(nil)

type HttpGetter func(urlPath string) ([]byte, string, error)

// RemoteMD represents a remote Markdown file.
// Main steps to get a TOC:
// - download it locally
// - convert md file to HTML (via GH API)
// - grab TOC from HTML
type RemoteMD struct {
	LocalMD
	HttpGet HttpGetter
}

func (md *RemoteMD) download() error {
	htmlBody, ContentType, err := md.HttpGet(md.Path)
	md.Log("try to download remote file=%s. content-type=%s", md.Path, ContentType)
	if err != nil {
		md.Log("Download err: %s", err)
		return err
	}

	// if not a plain text - it's an error
	if strings.Split(ContentType, ";")[0] != "text/plain" {
		md.Log("not a plain text, stop.")
		return err
	}

	// if remote file's content is a plain text
	// we need to convert it to html
	tmpfile, err := os.CreateTemp("", "ghtoc-remote-txt-*")
	if err != nil {
		md.Log("Creating file err: %s", err)
		return err
	}
	defer tmpfile.Close()

	md.Path = tmpfile.Name()
	md.Log("save content into: %s", md.Path)
	if err = os.WriteFile(tmpfile.Name(), htmlBody, 0644); err != nil {
		md.Log("Writing file err: %s", err)
		return err
	}
	return nil
}

// GetToc downloads file, converts it into HTML & grab TOC.
func (md *RemoteMD) GetToc() ghtoc.GHToc {
	if err := md.download(); err != nil {
		md.Log("Error while downloading file: %s", err)
		return nil
	}

	return md.LocalMD.GetToc()
}
