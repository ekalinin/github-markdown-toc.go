package file

import (
	"os"

	ghtoc "github.com/ekalinin/github-markdown-toc.go"
	"github.com/ekalinin/github-markdown-toc.go/internal/grabber"
)

var _ Tocer = (*LocalMD)(nil)

// RemoteHTML represents a remote HTML file.
// Main steps to get a TOC:
// - download JSON locally (yep, GH returns JSON now)
// - get TOC from JSON (there's a special section)
type RemoteHTML struct {
	file
	HttpGet    HttpGetter
	tocGrabber grabber.Grabber
}

func (md *RemoteHTML) Toc() ghtoc.GHToc {
	md.Log("downloading remote file=%s ...", md.Path)
	jsonBody, ContentType, err := md.HttpGet(md.Path)
	md.Log(" ... remote file. content-type=%s", ContentType)
	if err != nil {
		md.Log(" ... download err=" + err.Error())
		return nil
	}

	if md.Debug {
		tmpfile, err := os.CreateTemp("", "ghtoc-remote-json-*")
		if err != nil {
			md.Log("creating file err: %s", err)
			return nil
		}
		defer tmpfile.Close()
		md.Path = tmpfile.Name()

		jsonFile := md.Path + ".debug.json"
		md.Log("writing json file: %s", jsonFile)
		if err := os.WriteFile(jsonFile, jsonBody, 0644); err != nil {
			md.Log("writing json file error: %s", err)
			return nil
		}
	}

	md.Log("grabbing the TOC ...")
	toc, err := md.tocGrabber.Grab(string(jsonBody))
	if err != nil {
		md.Log("failed to grab TOC: %s", err)
		return nil
	}

	md.Log("done.")
	return toc
}
