package ghtoc

import (
	"io"
	"log"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// HxDepth represents the header depth with H1 being 0.
type HxDepth int

// InvalidDepth designates that the data atom is not a valid Hx.
const InvalidDepth HxDepth = -1

const _headerRegexpStr = `(?si)<h(?P<num>[1-6])>\s*` +
	`<a\s*id="user-content-[^"]*"\s*class="anchor"\s*` +
	`href="(?P<href>[^"]*)"[^>]*>\s*` +
	`.*?</a>(?P<name>.*?)</h`

// const _headerRegexpStr = `(?si)<h(?P<num>[1-6])>\s*` +
// 	`<a\s.*` +
// 	`\bid="user-content-[^"]*"\s.*` +
// 	`\bclass="anchor"\s.*` +
// 	`\bhref="(?P<href>[^"]*)"\s.*` +
// 	`[^>]*>` +
// 	`.*?</a>(?P<name>.*?)</h`

// const _newHeaderRegexpStr = `(?si)<h(?P<num>[1-6])>\s*` +
// 	`<a\s.*\bid="user-content-[^"]*"\s.*\bclass="anchor"\s.*` +
// 	`\bhref="(?P<href>[^"]*)"[^>]*>\s*` +
// 	`.*?</a>(?P<name>.*?)</h`

var _headerRegexp *regexp.Regexp

// var _newHeaderRegexp *regexp.Regexp

func headerRegexp() *regexp.Regexp {
	if _headerRegexp == nil {
		_headerRegexp = regexp.MustCompile(_headerRegexpStr)
	}
	return _headerRegexp
}

// func newHeaderRegexp() *regexp.Regexp {
// 	if _newHeaderRegexp == nil {
// 		_newHeaderRegexp = regexp.MustCompile(_newHeaderRegexpStr)
// 	}
// 	return _newHeaderRegexp
// }

// Header represents an HTML header
type Header struct {
	Depth HxDepth
	Href  string
	Name  string
}

func findHeadersInString(str string) []*Header {
	r := strings.NewReader(str)
	return findHeaders(r)
}

func findHeaders(r io.Reader) []*Header {
	hdrs := make([]*Header, 0)
	tokenizer := html.NewTokenizer(r)
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			// TODO(chuck): Check if this is io.EOF?
			return hdrs
		case html.StartTagToken:
			t := tokenizer.Token()

			// DEBUG BEGIN
			log.Printf("*** CHUCK:  default t: %+#v", t)
			// log.Printf("*** CHUCK: default t.Type: %+#v", t.Type)
			// log.Printf("*** CHUCK: default t.DataAtom: %+#v", t.DataAtom)
			// DEBUG END

			hdr := createHeader(tokenizer, t)
			if hdr != nil {
				hdrs = append(hdrs, hdr)
			}
		}
	}
}

// func isHxTag(dataAtom atom.Atom) bool {
// 	depth := getHxDepth(dataAtom)
// 	return (depth != InvalidDepth)
// }

func getHxDepth(dataAtom atom.Atom) HxDepth {
	depths := []atom.Atom{
		atom.H1,
		atom.H2,
		atom.H3,
		atom.H4,
		atom.H5,
		atom.H6,
	}
	for depth, v := range depths {
		if dataAtom == v {
			return HxDepth(depth)
		}
	}
	return InvalidDepth
}

func createHeader(tokenizer *html.Tokenizer, token html.Token) *Header {
	hxDepth := getHxDepth(token.DataAtom)
	if hxDepth == InvalidDepth {
		return nil
	}

	var href, name string
	tokenDepth := 0
	for {
		tokenizer.Next()
		t := tokenizer.Token()
		switch t.Type {
		case html.ErrorToken:
			return nil
		case html.StartTagToken:
			tokenDepth++
			if t.DataAtom == atom.A {
				if hrefAttr, ok := findAttribute(t.Attr, "", "href"); ok {
					href = hrefAttr.Val
				} else {
					// Expected to find href attribute
					return nil
				}
			}
		case html.EndTagToken:
			// If we encountered the matching end tag for the Hx, then we are done
			if t.DataAtom == token.DataAtom {
				return &Header{
					Depth: hxDepth,
					Name:  name,
					Href:  href,
				}
			}
			tokenDepth--
		case html.TextToken:
			if tokenDepth == 1 {
				name = t.Data
			}
		}
	}
}

func findAttribute(attrs []html.Attribute, namespace, key string) (html.Attribute, bool) {
	for _, attr := range attrs {
		if attr.Namespace == namespace && attr.Key == key {
			return attr, true
		}
	}
	return html.Attribute{}, false
}
