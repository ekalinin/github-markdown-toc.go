package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	version   = "0.7.0"
	userAgent = fmt.Sprint("github-markdown-toc.go v", version)
)

// GHToc GitHub TOC
type GHToc []string

// Print print TOC to the console
func (toc *GHToc) Print() {
	for _, tocItem := range *toc {
		fmt.Println(tocItem)
	}
	fmt.Println()
}

// GHDoc GitHub document
type GHDoc struct {
	Path     string
	AbsPaths bool
	Depth    int
	Escape   bool
}

// NewGHDoc create GHDoc
func NewGHDoc(Path string, AbsPaths bool, Depth int, Escape bool) *GHDoc {
	return &GHDoc{Path, AbsPaths, Depth, Escape}
}

// GetToc return GHToc for a document
func (doc *GHDoc) GetToc() *GHToc {
	htmlBody, err := GetHmtlBody(doc.Path)
	if err != nil {
		return nil
	}
	if doc.AbsPaths {
		return GrabTocX(htmlBody, doc.Path, doc.Depth, doc.Escape)
	}
	return GrabTocX(htmlBody, "", doc.Depth, doc.Escape)
}

//
// Internal
//

// check checks if there whas an error and do panic if it was
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// doHTTPReq executes a particullar http request
func doHTTPReq(req *http.Request) (string, error) {
	req.Header.Set("User-Agent", userAgent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Executes HTTP GET request
func httpGet(urlPath string) (string, error) {
	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return "", err
	}
	return doHTTPReq(req)
}

// httpPost executes HTTP POST with file content
func httpPost(urlPath string, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	io.Copy(body, file)

	req, err := http.NewRequest("POST", urlPath, body)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "text/plain")

	return doHTTPReq(req)
}

// removeStuf trims spaces, removes new lines and code tag from a string
func removeStuf(s string) string {
	res := strings.Replace(s, "\n", "", -1)
	res = strings.Replace(res, "<code>", "", -1)
	res = strings.Replace(res, "</code>", "", -1)
	res = strings.TrimSpace(res)

	return res
}

// Public

// EscapeSpecChars Escapes special characters
func EscapeSpecChars(s string) string {
	specChar := []string{"\\", "`", "*", "_", "{", "}", "#", "+", "-", ".", "!"}
	res := s

	for _, c := range specChar {
		res = strings.Replace(res, c, "\\"+c, -1)
	}
	return res
}

// GetHmtlBody If path is url then just executes HTTP GET and
// Returns html for this url.
//
// If path is a local path then sends file to the GitHub's
// Markdown -> Html converter and returns html.
func GetHmtlBody(path string) (string, error) {
	if IsURL(path) {
		return httpGet(path)
	}
	return ConvertMd2Html(path)
}

// IsURL Check if string is url
func IsURL(candidate string) bool {
	u, err := url.Parse(candidate)
	if err != nil || u.Scheme == "" {
		return false
	}
	return true
}

// ConvertMd2Html Sends Markdown to the github converter
// and returns html
func ConvertMd2Html(localpath string) (string, error) {
	return httpPost("https://api.github.com/markdown/raw", localpath)
}

// GrabToc Create TOC by html from github
func GrabToc(html string, absPath string, depth int) *GHToc {
	return GrabTocX(html, absPath, depth, true)
}

// GrabTocX Create TOC
func GrabTocX(html string, absPath string, depth int, escape bool) *GHToc {
	re := `(?si)<h(?P<num>[1-6])>\s*` +
		`<a\s*id="user-content-[^"]*"\s*class="anchor"\s*` +
		`href="(?P<href>[^"]*)"[^>]*>\s*` +
		`.*?</a>(?P<name>.*?)</h`
	r := regexp.MustCompile(re)

	toc := GHToc{}
	minHeaderNum := 6
	var groups []map[string]string
	for _, match := range r.FindAllStringSubmatch(html, -1) {
		group := make(map[string]string)
		// fill map for groups
		for i, name := range r.SubexpNames() {
			if i == 0 || name == "" {
				continue
			}
			group[name] = removeStuf(match[i])
		}
		// update minimum header number
		n, _ := strconv.Atoi(group["num"])
		if n < minHeaderNum {
			minHeaderNum = n
		}
		groups = append(groups, group)
	}

	var tmpSection string
	for _, group := range groups {
		// format result
		n, _ := strconv.Atoi(group["num"])
		if depth > 0 && n > depth {
			continue
		}

		link := group["href"]
		if len(absPath) > 0 {
			link = absPath + link
		}

		tmpSection = removeStuf(group["name"])
		if escape {
			tmpSection = EscapeSpecChars(tmpSection)
		}
		tocItem := strings.Repeat("  ", n-minHeaderNum) + "* " +
			"[" + tmpSection + "]" +
			"(" + link + ")"
		//fmt.Println(tocItem)
		toc = append(toc, tocItem)
	}

	return &toc
}

// Entry point
func main() {
	pathsDesc := "Local path or URL of the document to grab TOC. " +
		"If not entered, then read Markdown from stdin."
	paths := kingpin.Arg("path", pathsDesc).Strings()
	serial := kingpin.Flag("serial", "Grab TOCs in the serial mode").Bool()
	hideHeader := kingpin.Flag("hide-header", "Hide TOC header").Bool()
	depth := kingpin.Flag("depth",
		"How many levels of headings to include. Defaults to 0 (all)").Default("0").Int()
	noEscape := kingpin.Flag("no-escape", "Do not escape chars in sections").Bool()
	kingpin.Version(version)
	kingpin.Parse()

	pathsCount := len(*paths)

	if !*hideHeader && pathsCount == 1 {
		fmt.Println()
		fmt.Println("Table of Contents")
		fmt.Println("=================")
		fmt.Println()
	}

	// read file paths | urls from args
	absPathsInToc := pathsCount > 1
	ch := make(chan *GHToc, pathsCount)

	for _, p := range *paths {
		ghdoc := NewGHDoc(p, absPathsInToc, *depth, !*noEscape)
		if *serial {
			ch <- ghdoc.GetToc()
		} else {
			go func(path string) { ch <- ghdoc.GetToc() }(p)
		}
	}

	for i := 1; i <= pathsCount; i++ {
		toc := <-ch
		toc.Print()
	}

	// read md from stdin
	if pathsCount == 0 {
		bytes, err := ioutil.ReadAll(os.Stdin)
		check(err)

		file, err := ioutil.TempFile(os.TempDir(), "ghtoc")
		check(err)
		defer os.Remove(file.Name())

		check(ioutil.WriteFile(file.Name(), bytes, 0644))
		NewGHDoc(file.Name(), false, *depth, !*noEscape).GetToc().Print()
	}

	fmt.Println("Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)")
}
