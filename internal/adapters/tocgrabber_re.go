package adapters

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/core/entity"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/utils"
	"github.com/ekalinin/github-markdown-toc.go/v2/internal/version"
)

type ReGrabber struct {
	cfg GrabberCfg

	re *regexp.Regexp
}

func NewReGrabber(path string, cfg GrabberCfg, reVersion string) *ReGrabber {
	// si:
	// 	- s - let . match \n (single-line mode)
	//  - i - case-insensitive
	re := ""
	if reVersion == version.GH_V0 {
		re = `(?si)<h(?P<num>[1-6])>\s*` +
			`<a\s*id="user-content-[^"]*"\s*class="anchor"\s*` +
			`(aria-hidden="[^"]*"\s*)?` +
			`(tabindex="[^"]*"\s*)?` +
			`href="(?P<href>[^"]*)"[^>]*>\s*` +
			`.*?</a>(?P<name>.*?)</h`
	}
	if reVersion == version.GH_2023_10 {
		re = `(?si)<h(?P<num>[1-6]) id="[^"]+">\s*` +
			`<a class="heading-link"\s*` +
			`href="(?P<href>[^"]+)">\s*` +
			`(?P<name>.*?)<span`
	}
	if reVersion == version.GH_2024_03 {
		re = `(?si)<h(?P<num>[1-6]) class="heading-element">(?P<name>.*?)</h\d>` +
			`<a\s*id="user-content-[^"]*"\s*` +
			`class="[^"]*"\s*` +
			`aria-label="[^"]*"\s*` +
			`href="(?P<href>[^"]+)">`
	}

	return &ReGrabber{
		cfg: cfg,
		re:  regexp.MustCompile(re),
	}
}

func (g *ReGrabber) Grab(html string) (*entity.Toc, error) {

	listIndentation := utils.GenerateListIndentation(g.cfg.Indent)

	toc := entity.Toc{}
	minHeaderNum := 6
	var groups []map[string]string
	// doc.d("GrabToc: matching ...")
	for _, match := range g.re.FindAllStringSubmatch(html, -1) {
		// doc.d("GrabToc: match #" + strconv.Itoa(idx) + " ...")
		group := make(map[string]string)
		// fill map for groups
		for i, name := range g.re.SubexpNames() {
			if i == 0 || name == "" {
				continue
			}
			// doc.d("GrabToc: process group: " + name + ": " + match[i] + " ...")
			group[name] = utils.RemoveStuff(match[i])
		}
		// update minimum header number
		n, _ := strconv.Atoi(group["num"])
		if n < minHeaderNum {
			minHeaderNum = n
		}
		groups = append(groups, group)
	}

	var tmpSection string
	// doc.d("GrabToc: processing groups ...")
	// doc.d("Including starting from level " + strconv.Itoa(doc.StartDepth))
	for _, group := range groups {
		// format result
		n, _ := strconv.Atoi(group["num"])
		if n <= g.cfg.StartDepth {
			continue
		}
		if g.cfg.Depth > 0 && n > g.cfg.Depth {
			continue
		}

		link, _ := url.QueryUnescape(group["href"])
		if g.cfg.AbsPaths {
			link = g.cfg.Path + link
		}

		tmpSection = utils.RemoveStuff(group["name"])
		if g.cfg.Escape {
			tmpSection = utils.EscapeSpecChars(tmpSection)
		}
		tocItem := strings.Repeat(listIndentation(), n-minHeaderNum-g.cfg.StartDepth) + "* " +
			"[" + tmpSection + "]" +
			"(" + link + ")"
		//fmt.Println(tocItem)
		toc = append(toc, tocItem)
	}

	return &toc, nil
}
