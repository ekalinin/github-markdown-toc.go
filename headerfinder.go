package ghtoc

import (
	"io"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// HxDepth represents the header depth with H1 being 0.
type HxDepth int

// InvalidDepth designates that the data atom is not a valid Hx.
const InvalidDepth HxDepth = -1

// MaxHxDepth is the maximum HxDepth value.
// H6 is the last Hx tag (5 = 6 - 1)
const MaxHxDepth HxDepth = 5

// Header represents an HTML header
type Header struct {
	Depth HxDepth
	Href  string
	Name  string
}

func findHeadersInString(str string) []Header {
	r := strings.NewReader(str)
	return findHeaders(r)
}

func findHeaders(r io.Reader) []Header {
	hdrs := make([]Header, 0)
	tokenizer := html.NewTokenizer(r)
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return hdrs
		case html.StartTagToken:
			t := tokenizer.Token()
			if hdr, ok := createHeader(tokenizer, t); ok {
				hdrs = append(hdrs, hdr)
			}
		}
	}
}

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

func createHeader(tokenizer *html.Tokenizer, token html.Token) (Header, bool) {
	hxDepth := getHxDepth(token.DataAtom)
	if hxDepth == InvalidDepth {
		return Header{}, false
	}

	var href string
	var nameParts []string
	// Start at 1 because we are inside the Hx tag
	tokenDepth := 1
	afterAnchor := false
	for {
		tokenizer.Next()
		t := tokenizer.Token()
		switch t.Type {
		case html.ErrorToken:
			return Header{}, false
		case html.StartTagToken:
			tokenDepth++
			if t.DataAtom == atom.A {
				if hrefAttr, ok := findAttribute(t.Attr, "", "href"); ok {
					href = hrefAttr.Val
				} else {
					// Expected to find href attribute
					return Header{}, false
				}
			}
		case html.EndTagToken:
			switch t.DataAtom {
			case token.DataAtom:
				// If we encountered the matching end tag for the Hx, then we are done
				return Header{
					Depth: hxDepth,
					Name:  removeStuff(strings.Join(nameParts, " ")),
					Href:  href,
				}, true
			case atom.A:
				afterAnchor = true
			}
			tokenDepth--
		case html.TextToken:
			if afterAnchor {
				nameParts = append(nameParts, removeStuff(t.Data))
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
