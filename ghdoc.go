package ghtoc

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/ekalinin/github-markdown-toc.go/internal"
)

// Print TOC to the console
func (toc *GHToc) Print(w io.Writer) error {
	for _, tocItem := range *toc {
		if _, err := fmt.Fprintln(w, tocItem); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	return nil
}

type httpGetter func(urlPath string) ([]byte, string, error)
type httpPoster func(urlPath, filePath, token string) (string, error)

// GHDoc GitHub document
type GHDoc struct {
	Path       string
	AbsPaths   bool
	StartDepth int
	Depth      int
	Escape     bool
	GhToken    string
	Indent     int
	Debug      bool

	// internals
	html       string
	logger     *log.Logger
	httpGetter httpGetter
	httpPoster httpPoster
	ghURL      string
}

// NewGHDoc create GHDoc
func NewGHDoc(Path string, AbsPaths bool, StartDepth int, Depth int, Escape bool, Token string, Indent int, Debug bool) *GHDoc {
	return &GHDoc{
		Path:       Path,
		AbsPaths:   AbsPaths,
		StartDepth: StartDepth,
		Depth:      Depth,
		Escape:     Escape,
		GhToken:    Token,
		Indent:     Indent,
		Debug:      Debug,
		html:       "",
		logger:     log.New(os.Stderr, "", log.LstdFlags),
		httpGetter: internal.HttpGet,
		httpPoster: internal.HttpPost,
		ghURL:      "https://api.github.com",
	}
}

func (doc *GHDoc) d(msg string) {
	if doc.Debug {
		doc.logger.Println(msg)
	}
}

// SetGHURL sets new GitHub URL (protocol + host)
func (doc *GHDoc) SetGHURL(url string) *GHDoc {
	if url != "" {
		doc.ghURL = url
	}
	return doc
}

// IsRemoteFile checks if path is for remote file or not
func (doc *GHDoc) IsRemoteFile() bool {
	u, err := url.Parse(doc.Path)
	if err != nil || u.Scheme == "" {
		doc.d("IsRemoteFile: false")
		return false
	}
	doc.d("IsRemoteFile: true")
	return true
}

func (doc *GHDoc) convertMd2Html(localPath string, token string) (string, error) {
	ghURL := doc.ghURL + "/markdown/raw"
	return doc.httpPoster(ghURL, localPath, token)
}

// Convert2HTML downloads remote file
func (doc *GHDoc) Convert2HTML() error {
	doc.d("Convert2HTML: start.")
	defer doc.d("Convert2HTML: done.")

	// remote file may be of 2 types:
	// - raw md file (we need to download it locally and convert t HTML)
	// - html file (we need just to load it and parse TOC from it)
	if doc.IsRemoteFile() {
		htmlBody, ContentType, err := doc.httpGetter(doc.Path)
		doc.d("Convert2HTML: remote file. content-type: " + ContentType)
		if err != nil {
			doc.d("Convert2HTML: err=" + err.Error())
			return err
		}

		// if not a plain text - return the result (should be html)
		if strings.Split(ContentType, ";")[0] != "text/plain" {
			doc.html = string(htmlBody)
			doc.d("Convert2HTML: not a plain text, body")
			return nil
		}

		// if remote file's content is a plain text
		// we need to convert it to html
		tmpfile, err := os.CreateTemp("", "ghtoc-remote-txt")
		if err != nil {
			return err
		}
		defer tmpfile.Close()
		doc.Path = tmpfile.Name()
		if err = os.WriteFile(tmpfile.Name(), htmlBody, 0644); err != nil {
			return err
		}
	}
	doc.d("Convert2HTML: local file: " + doc.Path)
	if _, err := os.Stat(doc.Path); os.IsNotExist(err) {
		return err
	}
	htmlBody, err := doc.convertMd2Html(doc.Path, doc.GhToken)
	doc.d("Convert2HTML: converted to html, size: " + strconv.Itoa(len(htmlBody)))
	if err != nil {
		return err
	}
	if doc.Debug {
		htmlFile := doc.Path + ".debug.html"
		doc.d("Convert2HTML: write html file: " + htmlFile)
		if err := os.WriteFile(htmlFile, []byte(htmlBody), 0644); err != nil {
			return err
		}
	}
	doc.html = htmlBody
	return nil
}

// GrabToc gets TOC from html
func (doc *GHDoc) GrabToc() *GHToc {
	doc.d("GrabToc: start, html size: " + strconv.Itoa(len(doc.html)))
	defer doc.d("GrabToc: done.")

	// si:
	// 	- s - let . match \n (single-line mode)
	//  - i - case-insensitive
	re := `(?si)<h(?P<num>[1-6]) id="[^"]+">\s*` +
		`<a class="heading-link"\s*` +
		`href="(?P<href>[^"]+)">\s*` +
		`(?P<name>.*?)<span`
	r := regexp.MustCompile(re)
	listIndentation := internal.GenerateListIndentation(doc.Indent)

	toc := GHToc{}
	minHeaderNum := 6
	var groups []map[string]string
	doc.d("GrabToc: matching ...")
	for idx, match := range r.FindAllStringSubmatch(doc.html, -1) {
		doc.d("GrabToc: match #" + strconv.Itoa(idx) + " ...")
		group := make(map[string]string)
		// fill map for groups
		for i, name := range r.SubexpNames() {
			if i == 0 || name == "" {
				continue
			}
			doc.d("GrabToc: process group: " + name + ": " + match[i] + " ...")
			group[name] = internal.RemoveStuff(match[i])
		}
		// update minimum header number
		n, _ := strconv.Atoi(group["num"])
		if n < minHeaderNum {
			minHeaderNum = n
		}
		groups = append(groups, group)
	}

	var tmpSection string
	doc.d("GrabToc: processing groups ...")
	doc.d("Including starting from level " + strconv.Itoa(doc.StartDepth))
	for _, group := range groups {
		// format result
		n, _ := strconv.Atoi(group["num"])
		if n <= doc.StartDepth {
			continue
		}
		if doc.Depth > 0 && n > doc.Depth {
			continue
		}

		link, _ := url.QueryUnescape(group["href"])
		if doc.AbsPaths {
			link = doc.Path + link
		}

		tmpSection = internal.RemoveStuff(group["name"])
		if doc.Escape {
			tmpSection = internal.EscapeSpecChars(tmpSection)
		}
		tocItem := strings.Repeat(listIndentation(), n-minHeaderNum-doc.StartDepth) + "* " +
			"[" + tmpSection + "]" +
			"(" + link + ")"
		//fmt.Println(tocItem)
		toc = append(toc, tocItem)
	}

	return &toc
}

// GetToc return GHToc for a document
func (doc *GHDoc) GetToc() *GHToc {
	if err := doc.Convert2HTML(); err != nil {
		log.Fatal(err)
		return nil
	}
	return doc.GrabToc()
}
