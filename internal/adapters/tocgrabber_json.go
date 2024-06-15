package adapters

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/ekalinin/github-markdown-toc.go/internal/core/entity"
	"github.com/ekalinin/github-markdown-toc.go/internal/utils"
)

type defaultGrabber struct {
	Path string

	// toc grabber
	AbsPaths   bool
	StartDepth int
	Depth      int
	Escape     bool
	Indent     int
}

// ------------------------------------------------------
//

type JsonGrabber struct {
	defaultGrabber
}

func NewJsonGrabber(path string) *JsonGrabber {
	return &JsonGrabber{
		defaultGrabber: defaultGrabber{
			Path:       path,
			AbsPaths:   false,
			StartDepth: 0,
			Depth:      0,
			Escape:     true,
			Indent:     2,
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

func (g JsonGrabber) Grab(jsonBody string) (*entity.Toc, error) {
	var wrapper tocWrapper
	err := json.Unmarshal([]byte(jsonBody), &wrapper)
	if err != nil {
		return nil, err
	}

	// g.Log("processing groups ...")

	toc := entity.Toc{}
	tmpSection := ""
	listIndentation := utils.GenerateListIndentation(g.Indent)
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

		link, _ := url.QueryUnescape(item.Anchor)
		link = "#" + link
		// if err != nil {
		// 	g.Log("got error from query unescape: ", err.Error())
		// }
		if g.AbsPaths {
			link = g.Path + link
		}
		tmpSection = utils.RemoveStuff(item.Text)
		if g.Escape {
			tmpSection = utils.EscapeSpecChars(tmpSection)
		}

		prefix := strings.Repeat(listIndentation(), item.Level-minHeaderNum-g.StartDepth)
		tocItem := prefix + "* " +
			"[" + tmpSection + "]" +
			"(" + link + ")"
		toc = append(toc, tocItem)
	}

	return &toc, nil
}
