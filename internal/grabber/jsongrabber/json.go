package jsongrabber

import (
	"encoding/json"
	"net/url"
	"strings"

	ghtoc "github.com/ekalinin/github-markdown-toc.go"
	"github.com/ekalinin/github-markdown-toc.go/internal"
	"github.com/ekalinin/github-markdown-toc.go/internal/debugger"
	"github.com/ekalinin/github-markdown-toc.go/internal/grabber"
)

type JsonGrabber struct {
	grabber.DefaultGrabber
}

func New(path string) JsonGrabber {
	return JsonGrabber{
		DefaultGrabber: grabber.DefaultGrabber{
			Path:       path,
			AbsPaths:   false,
			StartDepth: 0,
			Depth:      0,
			Escape:     true,
			Indent:     2,
			Debugger:   debugger.New(false, "JsonGrabber"),
		},
	}
}

type tocItem struct {
	Level  int
	Text   string
	Anchor string
}

type tocWrapper struct {
	Payload struct {
		Blob struct {
			HeaderInfo struct {
				Toc []tocItem
			}
		}
	}
}

func (g JsonGrabber) Grab(jsonBody string) (ghtoc.GHToc, error) {
	var wrapper tocWrapper
	g.Log("matching ...")
	err := json.Unmarshal([]byte(jsonBody), &wrapper)
	if err != nil {
		return nil, err
	}

	g.Log("processing groups ...")

	toc := ghtoc.GHToc{}
	tmpSection := ""
	listIndentation := internal.GenerateListIndentation(g.Indent)
	minHeaderNum := 6
	for _, item := range wrapper.Payload.Blob.HeaderInfo.Toc {
		if item.Level < minHeaderNum {
			minHeaderNum = item.Level
		}
	}
	for _, item := range wrapper.Payload.Blob.HeaderInfo.Toc {
		if item.Level <= g.StartDepth {
			continue
		}
		if g.Depth > 0 && item.Level > g.Depth {
			continue
		}

		link, err := url.QueryUnescape(item.Anchor)
		link = "#" + link
		if err != nil {
			g.Log("got error from query unescape: ", err.Error())
		}
		if g.AbsPaths {
			link = g.Path + link
		}
		tmpSection = internal.RemoveStuff(item.Text)
		if g.Escape {
			tmpSection = internal.EscapeSpecChars(tmpSection)
		}

		prefix := strings.Repeat(listIndentation(), item.Level-minHeaderNum-g.StartDepth)
		tocItem := prefix + "* " +
			"[" + tmpSection + "]" +
			"(" + link + ")"
		toc = append(toc, tocItem)
	}

	return toc, nil
}
