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
	"regexp"
	"strconv"
	"strings"
)

var (
	version    = "0.2.1"
	user_agent = fmt.Sprint("github-markdown-toc.go v", version)
)

//
// Internal
//

func do_http_req(req *http.Request) string {
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
func http_get(url_path string) string {
	req, err := http.NewRequest("GET", url_path, nil)
	if err != nil {
		return ""
	}
	return do_http_req(req)
}

// Executes HTTP POST with file content
func http_post(url_path string, file_path string) string {
	file, err := os.Open(file_path)
	if err != nil {
		return ""
	}
	defer file.Close()

	body := &bytes.Buffer{}
	io.Copy(body, file)

	req, err := http.NewRequest("POST", url_path, body)
	req.Header.Set("Content-Type", "text/plain")

	return do_http_req(req)
}

// Public

// If path is url then just executes HTTP GET and
// Returns html for this url.
//
// If path is a local path then sends file to the GitHub's
// Markdown -> Html converter and returns html.
func GetHmtlBody(path string) string {
	if IsUrl(path) {
		return http_get(path)
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
	return http_post("https://api.github.com/markdown/raw", localpath)
}

// Create TOC by html from github
func GrabToc(html string) []string {
	re := `(?si)<h(?P<num>[1-6])>\s*` +
		`<a\s*id="user-content-[^"]*"\s*class="anchor"\s*` +
		`href="(?P<href>[^"]*)"[^>]*>\s*` +
		`<span[^<*]*</span>\s*</a>(?P<name>[^<]*)`
	r := regexp.MustCompile(re)

	toc := []string{}
	groups := make(map[string]string)
	for _, match := range r.FindAllStringSubmatch(html, -1) {
		// fill map for groups
		for i, name := range r.SubexpNames() {
			if i == 0 || name == "" {
				continue
			}
			groups[name] = strings.TrimSpace(
				strings.Replace(match[i], "\n", "", -1))
		}
		// format result
		n, _ := strconv.Atoi(groups["num"])
		toc_item := strings.Repeat("  ", n) + "* " +
			"[" + groups["name"] + "]" +
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

// Entry point
func main() {
	paths := kingpin.Arg("path", "Local path or URL of the document to grab TOC").Strings()
	kingpin.Version(version)
	kingpin.Parse()

	if len(*paths) == 1 {
		fmt.Println()
		fmt.Println("Table of Contents")
		fmt.Println("=================")
		fmt.Println()
	}

	for _, p := range *paths {
		for _, item := range GenerateToc(p) {
			fmt.Println(item)
		}
		fmt.Println()
	}
	fmt.Println()
	fmt.Println("Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)")
}
