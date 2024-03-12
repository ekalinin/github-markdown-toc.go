package regrabber

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	ghtoc "github.com/ekalinin/github-markdown-toc.go"
	"github.com/ekalinin/github-markdown-toc.go/internal"
	"github.com/ekalinin/github-markdown-toc.go/internal/debugger"
	"github.com/ekalinin/github-markdown-toc.go/internal/grabber"
)

type ReGrabber struct {
	grabber.DefaultGrabber

	// internals
	re *regexp.Regexp
}

func New(path string) ReGrabber {
	// si:
	// 	- s - let . match \n (single-line mode)
	//  - i - case-insensitive
	re := `(?si)<h(?P<num>[1-6]) id="[^"]+">\s*` +
		`<a class="heading-link"\s*` +
		`href="(?P<href>[^"]+)">\s*` +
		`(?P<name>.*?)<span`

	return ReGrabber{
		DefaultGrabber: grabber.DefaultGrabber{
			Path:       path,
			AbsPaths:   false,
			StartDepth: 0,
			Depth:      0,
			Escape:     true,
			Indent:     2,
			Debugger:   debugger.New(false, "ReGrabber"),
		},
		re: regexp.MustCompile(re),
	}
}

func (g ReGrabber) Grab(html string) (ghtoc.GHToc, error) {
	g.Log("matching ...")

	minHeaderNum := 6
	var groups []map[string]string
	for idx, match := range g.re.FindAllStringSubmatch(html, -1) {
		g.Log("match #%d ...\n", idx)
		group := make(map[string]string)
		// fill map for groups
		for i, name := range g.re.SubexpNames() {
			if i == 0 || name == "" {
				continue
			}
			g.Log("process group: %s: %s ...\n", name, match[i])
			group[name] = internal.RemoveStuff(match[i])
		}
		// update minimum header number
		n, err := strconv.Atoi(group["num"])
		if err != nil {
			g.Log("got error from atoi: ", err.Error())
		}
		if n < minHeaderNum {
			minHeaderNum = n
		}
		groups = append(groups, group)
	}

	g.Log("processing groups ...")

	toc := ghtoc.GHToc{}
	tmpSection := ""
	listIndentation := internal.GenerateListIndentation(g.Indent)
	g.Log("starting from level=%d\n", g.StartDepth)
	for _, group := range groups {
		// format result
		n, _ := strconv.Atoi(group["num"])
		if n <= g.StartDepth {
			continue
		}
		if g.Depth > 0 && n > g.Depth {
			continue
		}

		link, err := url.QueryUnescape(group["href"])
		if err != nil {
			g.Log("got error from query unescape: ", err.Error())
		}
		if g.AbsPaths {
			link = g.Path + link
		}

		tmpSection = internal.RemoveStuff(group["name"])
		if g.Escape {
			tmpSection = internal.EscapeSpecChars(tmpSection)
		}
		prefix := strings.Repeat(listIndentation(), n-minHeaderNum-g.StartDepth)
		tocItem := prefix + "* " +
			"[" + tmpSection + "]" +
			"(" + link + ")"
		toc = append(toc, tocItem)
	}

	return toc, nil
}
