package ghtoc

import "regexp"

// const _headerRegexpStr = `(?si)<h(?P<num>[1-6])>\s*` +
// 	`<a\s*id="user-content-[^"]*"\s*class="anchor"\s*` +
// 	`href="(?P<href>[^"]*)"[^>]*>\s*` +
// 	`.*?</a>(?P<name>.*?)</h`

const _headerRegexpStr = `(?si)<h(?P<num>[1-6])>\s*` +
	`<a\s.*` +
	`\bid="user-content-[^"]*"\s.*` +
	`\bclass="anchor"\s.*` +
	`\bhref="(?P<href>[^"]*)"\s.*` +
	`[^>]*>` +
	`.*?</a>(?P<name>.*?)</h`

const _newHeaderRegexpStr = `(?si)<h(?P<num>[1-6])>\s*` +
	`<a\s.*\bid="user-content-[^"]*"\s.*\bclass="anchor"\s.*` +
	`\bhref="(?P<href>[^"]*)"[^>]*>\s*` +
	`.*?</a>(?P<name>.*?)</h`

var _headerRegexp *regexp.Regexp
var _newHeaderRegexp *regexp.Regexp

func headerRegexp() *regexp.Regexp {
	if _headerRegexp == nil {
		_headerRegexp = regexp.MustCompile(_headerRegexpStr)
	}
	return _headerRegexp
}

func newHeaderRegexp() *regexp.Regexp {
	if _newHeaderRegexp == nil {
		_newHeaderRegexp = regexp.MustCompile(_newHeaderRegexpStr)
	}
	return _newHeaderRegexp
}
