package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	version = "0.1.0"
)

// Executes HTTP GET request
func http_get(url_path string) string {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url_path, nil)
	if err != nil {
		return ""
	}

	req.Header.Set("User-Agent", fmt.Sprint("ghtoc v", version))

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

// Check if string is url
func IsUrl(candidate string) bool {
	_, err := url.Parse(candidate)
	if err != nil {
		return false
	}
	return true
}

// Sends Markdown to the github converter
// and returns html.
func convert_md_to_html(localpath string) string {
	return ""
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

func get_hmtl(path string) string {
	if IsUrl(path) {
		return http_get(path)
	} else {
		return convert_md_to_html(path)
	}
}

// Generate TOC for document (path in filesystem or url)
func GenerateToc(path string) []string {
	return GrabToc(get_hmtl(path))
}

// Custom help
func usage() {
	app_name := strings.Replace(os.Args[0], "./", "", 1)

	fmt.Println("GitHub TOC generator: ", version)
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("	$", app_name, "url [url]")
	//fmt.Println("	$", app_name, "[options] [path [path]]")
	//fmt.Println("")
	//fmt.Println("Options:")

	flag.PrintDefaults()
}

// Entry point
func main() {

	flag.Usage = usage
	flag.Parse()

	paths := flag.Args()

	if len(paths) == 1 {
		fmt.Println()
		fmt.Println("Table of Contents")
		fmt.Println("=================")
		fmt.Println()
	}

	for _, p := range paths {
		for _, item := range GenerateToc(p) {
			fmt.Println(item)
		}
	}
	fmt.Println()
	fmt.Println("Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)")
}
