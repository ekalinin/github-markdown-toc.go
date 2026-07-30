package adapters

import (
	"context"
	"fmt"
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

func NewReGrabber(path string, cfg GrabberCfg, reVersion string) (*ReGrabber, error) {
	// si:
	// 	- s - let . match \n (single-line mode)
	//  - i - case-insensitive
	var pattern string
	switch reVersion {
	case version.GH_V0:
		pattern = `(?si)<h(?P<num>[1-6])>\s*` +
			`<a\s*id="user-content-[^"]*"\s*class="anchor"\s*` +
			`(aria-hidden="[^"]*"\s*)?` +
			`(tabindex="[^"]*"\s*)?` +
			`href="(?P<href>[^"]*)"[^>]*>\s*` +
			`.*?</a>(?P<name>.*?)</h`
	case version.GH_2023_10:
		pattern = `(?si)<h(?P<num>[1-6]) id="[^"]+">\s*` +
			`<a class="heading-link"\s*` +
			`href="(?P<href>[^"]+)">\s*` +
			`(?P<name>.*?)<span`
	case version.GH_2024_03:
		pattern = `(?si)<h(?P<num>[1-6]) class="heading-element">(?P<name>.*?)</h\d>` +
			`<a\s*id="user-content-[^"]*"\s*` +
			`class="[^"]*"\s*` +
			`aria-label="[^"]*"\s*` +
			`href="(?P<href>[^"]+)">`
	default:
		return nil, fmt.Errorf(
			"unsupported GitHub regexp version %q; supported versions: %s",
			reVersion,
			strings.Join(version.SupportedGHVersions(), ", "),
		)
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("compile GitHub regexp version %q: %w", reVersion, err)
	}

	return &ReGrabber{
		cfg: cfg,
		re:  re,
	}, nil
}

func (g *ReGrabber) Grab(ctx context.Context, html string) (*entity.Toc, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	listIndentation := utils.GenerateListIndentation(g.cfg.Indent)

	toc := entity.Toc{}
	minHeaderNum := 6
	var groups []map[string]string
	// doc.d("GrabToc: matching ...")
	for _, match := range g.re.FindAllStringSubmatch(html, -1) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
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
		if err := ctx.Err(); err != nil {
			return nil, err
		}
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
