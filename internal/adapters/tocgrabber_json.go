package adapters

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/ekalinin/github-markdown-toc.go/internal/core/entity"
	"github.com/ekalinin/github-markdown-toc.go/internal/utils"
)

type JsonGrabber struct {
	cfg GrabberCfg
}

func NewJsonGrabber(cfg GrabberCfg) *JsonGrabber {
	return &JsonGrabber{
		cfg: cfg,
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
		return nil, fmt.Errorf("got error from unmarshal: %w", err)
	}

	// g.Log("processing groups ...")

	toc := entity.Toc{}
	tmpSection := ""
	listIndentation := utils.GenerateListIndentation(g.cfg.Indent)
	minHeaderNum := 6
	for _, item := range wrapper.Payload.Blob.HeaderInfo.Toc {
		if item.Level < minHeaderNum {
			minHeaderNum = item.Level
		}
	}
	for _, item := range wrapper.Payload.Blob.HeaderInfo.Toc {
		if item.Level <= g.cfg.StartDepth {
			continue
		}
		if g.cfg.Depth > 0 && item.Level > g.cfg.Depth {
			continue
		}

		link, err := url.QueryUnescape(item.Anchor)
		if err != nil {
			// g.Log("got error from query unescape: ", err.Error())
			return nil, fmt.Errorf("got error from unescape: %w", err)
		}
		link = "#" + link
		if g.cfg.AbsPaths {
			link = g.cfg.Path + link
		}
		tmpSection = utils.RemoveStuff(item.Text)
		if g.cfg.Escape {
			tmpSection = utils.EscapeSpecChars(tmpSection)
		}

		prefix := strings.Repeat(listIndentation(), item.Level-minHeaderNum-g.cfg.StartDepth)
		tocItem := prefix + "* " +
			"[" + tmpSection + "]" +
			"(" + link + ")"
		toc = append(toc, tocItem)
	}

	return &toc, nil
}
