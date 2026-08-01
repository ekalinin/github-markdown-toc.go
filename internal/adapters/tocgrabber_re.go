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

// RegexpExtractor extracts headings from GitHub-rendered HTML.
type RegexpExtractor struct {
	re *regexp.Regexp
}

func NewRegexpExtractor(reVersion string) (*RegexpExtractor, error) {
	// si:
	//  - s - let . match \n (single-line mode)
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
	return &RegexpExtractor{re: re}, nil
}

func (e *RegexpExtractor) Extract(ctx context.Context, html string) ([]entity.Heading, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	matches := e.re.FindAllStringSubmatch(html, -1)
	headings := make([]entity.Heading, 0, len(matches))
	for _, match := range matches {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		group := make(map[string]string)
		for i, name := range e.re.SubexpNames() {
			if i != 0 && name != "" {
				group[name] = match[i]
			}
		}
		level, err := strconv.Atoi(group["num"])
		if err != nil {
			return nil, fmt.Errorf("parse GitHub heading level %q: %w", group["num"], err)
		}
		anchor, err := url.QueryUnescape(utils.RemoveStuff(group["href"]))
		if err != nil {
			return nil, fmt.Errorf("unescape GitHub heading anchor %q: %w", group["href"], err)
		}
		headings = append(headings, entity.Heading{
			Level:  level,
			Text:   utils.RemoveStuff(group["name"]),
			Anchor: strings.TrimPrefix(anchor, "#"),
		})
	}
	return headings, nil
}
