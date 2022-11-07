package ghtoc

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// GHToc GitHub TOC
type GHToc []string

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
	html       string
	logger     *log.Logger
	httpGetter httpGetter
	httpPoster httpPoster
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
		httpGetter: httpGet,
		httpPoster: httpPost,
	}
}

func (doc *GHDoc) d(msg string) {
	if doc.Debug {
		doc.logger.Println(msg)
	}
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
	ghURL := "https://api.github.com/markdown/raw"
	return doc.httpPoster(ghURL, localPath, token)
}

// Convert2HTML downloads remote file
func (doc *GHDoc) Convert2HTML() error {
	doc.d("Convert2HTML: start.")
	defer doc.d("Convert2HTML: done.")

	if doc.IsRemoteFile() {
		htmlBody, ContentType, err := doc.httpGetter(doc.Path)
		doc.d("Convert2HTML: remote file. content-type: " + ContentType)
		if err != nil {
			return err
		}

		// if not a plain text - return the result (should be html)
		if strings.Split(ContentType, ";")[0] != "text/plain" {
			doc.html = string(htmlBody)
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

	listIndentation := generateListIndentation(doc.Indent)

	minDepth := doc.StartDepth
	var maxDepth int
	if doc.Depth > 0 {
		maxDepth = doc.Depth - 1
	} else {
		maxDepth = int(MaxHxDepth)
	}

	toc := GHToc{}
	for _, hdr := range findHeadersInString(doc.html) {
		// DEBUG BEGIN
		log.Printf("*** CHUCK: GrabToc hdr: %+#v", hdr)
		log.Printf("*** CHUCK: GrabToc minDepth: %+#v", minDepth)
		log.Printf("*** CHUCK: GrabToc maxDepth: %+#v", maxDepth)
		// DEBUG END
		hDepth := int(hdr.Depth)
		if hDepth >= minDepth && hDepth <= maxDepth {
			toc = append(toc, doc.tocEntry(listIndentation(), hdr))
		}
	}

	return &toc
}

func (doc *GHDoc) tocEntry(indent string, hdr Header) string {
	// TODO(chuck): Calculate the repeat count with the doc.StartDepth
	return strings.Repeat(indent, int(hdr.Depth)) + "* " +
		"[" + doc.tocName(hdr.Name) + "]" +
		"(" + doc.tocLink(hdr.Href) + ")"
}

func (doc *GHDoc) tocName(name string) string {
	if doc.Escape {
		return EscapeSpecChars(name)
	}
	return name
}

func (doc *GHDoc) tocLink(href string) string {
	link, _ := url.QueryUnescape(href)
	if doc.AbsPaths {
		link = doc.Path + link
	}
	return link
}

//func (doc *GHDoc) GrabToc() *GHToc {
//	doc.d("GrabToc: start, html size: " + strconv.Itoa(len(doc.html)))
//	defer doc.d("GrabToc: done.")

//	r := headerRegexp()
//	listIndentation := generateListIndentation(doc.Indent)

//	toc := GHToc{}
//	minHeaderNum := 6
//	var groups []map[string]string
//	doc.d("GrabToc: matching ...")
//	for idx, match := range r.FindAllStringSubmatch(doc.html, -1) {
//		doc.d("GrabToc: match #" + strconv.Itoa(idx) + " ...")
//		group := make(map[string]string)
//		// fill map for groups
//		for i, name := range r.SubexpNames() {
//			if i == 0 || name == "" {
//				continue
//			}
//			doc.d("GrabToc: process group: " + name + ": " + match[i] + " ...")
//			group[name] = removeStuff(match[i])
//		}
//		// update minimum header number
//		n, _ := strconv.Atoi(group["num"])
//		if n < minHeaderNum {
//			minHeaderNum = n
//		}
//		groups = append(groups, group)
//	}

//	var tmpSection string
//	doc.d("GrabToc: processing groups ...")
//	doc.d("Including starting from level " + strconv.Itoa(doc.StartDepth))
//	for _, group := range groups {
//		// format result
//		n, _ := strconv.Atoi(group["num"])
//		if n <= doc.StartDepth {
//			continue
//		}
//		if doc.Depth > 0 && n > doc.Depth {
//			continue
//		}

//		link, _ := url.QueryUnescape(group["href"])
//		if doc.AbsPaths {
//			link = doc.Path + link
//		}

//		tmpSection = removeStuff(group["name"])
//		if doc.Escape {
//			tmpSection = EscapeSpecChars(tmpSection)
//		}
//		tocItem := strings.Repeat(listIndentation(), n-minHeaderNum-doc.StartDepth) + "* " +
//			"[" + tmpSection + "]" +
//			"(" + link + ")"
//		//fmt.Println(tocItem)
//		toc = append(toc, tocItem)
//	}

//	return &toc
//}

// GetToc return GHToc for a document
func (doc *GHDoc) GetToc() *GHToc {
	if err := doc.Convert2HTML(); err != nil {
		log.Fatal(err)
		return nil
	}
	return doc.GrabToc()
}
