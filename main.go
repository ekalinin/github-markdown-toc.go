package main

import (
	"bytes"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	version    = "0.3.1"
	user_agent = fmt.Sprint("github-markdown-toc.go v", version)
)

//
// Internal
//

// check checks if there whas an error and do panic if it was
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// doHttpReq executes a particullar http request
func doHttpReq(req *http.Request) string {
	req.Header.Set("User-Agent", user_agent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	return string(body)
}

// Executes HTTP GET request
func httpGet(url_path string) string {
	req, err := http.NewRequest("GET", url_path, nil)
	if err != nil {
		return ""
	}
	return doHttpReq(req)
}

// httpPost executes HTTP POST with file content
func httpPost(url_path string, file_path string) string {
	file, err := os.Open(file_path)
	if err != nil {
		return ""
	}
	defer file.Close()

	body := &bytes.Buffer{}
	io.Copy(body, file)

	req, err := http.NewRequest("POST", url_path, body)
	req.Header.Set("Content-Type", "text/plain")

	return doHttpReq(req)
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

// If path is url then just executes HTTP GET and
// Returns html for this url.
//
// If path is a local path then sends file to the GitHub's
// Markdown -> Html converter and returns html.
func GetHmtlBody(path string) string {
	if IsUrl(path) {
		return httpGet(path)
	} else {
		return ConvertMd2Html(path)
	}
}

// Check if string is url
func IsUrl(candidate string) bool {
	u, err := url.Parse(candidate)
	if err != nil || u.Scheme == "" {
		return false
	}
	return true
}

// Sends Markdown to the github converter
// and returns html.
func ConvertMd2Html(localpath string) string {
	return httpPost("https://api.github.com/markdown/raw", localpath)
}

// Create TOC by html from github
func GrabToc(html string) []string {
	re := `(?si)<h(?P<num>[1-6])>\s*` +
		`<a\s*id="user-content-[^"]*"\s*class="anchor"\s*` +
		`href="(?P<href>[^"]*)"[^>]*>\s*` +
		`<span[^<*]*</span>\s*</a>(?P<name>.*?)</h`
	r := regexp.MustCompile(re)

	toc := []string{}
	groups := make(map[string]string)
	for _, match := range r.FindAllStringSubmatch(html, -1) {
		// fill map for groups
		for i, name := range r.SubexpNames() {
			if i == 0 || name == "" {
				continue
			}
			groups[name] = removeStuf(match[i])
		}
		// format result
		n, _ := strconv.Atoi(groups["num"])
		toc_item := strings.Repeat("  ", n) + "* " +
			"[" + removeStuf(groups["name"]) + "]" +
			"(" + groups["href"] + ")"
		//fmt.Println(toc_item)
		toc = append(toc, toc_item)
	}

	return toc
}

// Generate TOC for document (path in filesystem or url)
func GenerateToc(path string) []string {
	return GrabToc(GetHmtlBody(path))
}

// PrintToc print on console string array
func PrintToc(toc []string) {
	for _, toc_item := range toc {
		fmt.Println(toc_item)
	}
	fmt.Println()
}

// Entry point
func main() {
	paths_desc := "Local path or URL of the document to grab TOC. " +
		"If not entered, then read Markdown from stdin."
	paths := kingpin.Arg("path", paths_desc).Strings()
	kingpin.Version(version)
	kingpin.Parse()

	if len(*paths) == 1 {
		fmt.Println()
		fmt.Println("Table of Contents")
		fmt.Println("=================")
		fmt.Println()
	}

	// read file paths | urls from args
	for _, p := range *paths {
		PrintToc(GenerateToc(p))
	}

	// read md from stdin
	if len(*paths) == 0 {
		file, err := ioutil.TempFile(os.TempDir(), "ghtoc")
		check(err)
		defer os.Remove(file.Name())
		file_path, err := filepath.Abs(file.Name())
		check(err)
		bytes, err := ioutil.ReadAll(os.Stdin)
		check(err)
		check(ioutil.WriteFile(file.Name(), bytes, 0644))
		PrintToc(GenerateToc(file_path))
	}

	fmt.Println("Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)")
}
