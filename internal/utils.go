package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

// doHTTPReq executes a particular http request
func doHTTPReq(req *http.Request) ([]byte, string, error) {
	req.Header.Set("User-Agent", userAgent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, "", err
	}

	if resp.StatusCode == http.StatusForbidden {
		return []byte{}, resp.Header.Get("Content-type"), errors.New(string(body))
	}

	return body, resp.Header.Get("Content-type"), nil
}

// HttpGet executes HTTP GET request.
func HttpGet(urlPath string) ([]byte, string, error) {
	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return []byte{}, "", err
	}
	return doHTTPReq(req)
}

func HttpGetJson(urlPath string) ([]byte, string, error) {
	req, err := http.NewRequest("GET", urlPath, nil)
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Accept", "application/json")
	if err != nil {
		return []byte{}, "", err
	}
	return doHTTPReq(req)
}

// HttpPost executes HTTP POST with file content.
func HttpPost(urlPath, filePath, token string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	_, err = io.Copy(body, file)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", urlPath, body)
	if err != nil {
		return "", err
	}

	if token != "" {
		req.Header.Add("Authorization", "token "+token)
	}
	req.Header.Set("Content-Type", "text/plain;charset=utf-8")

	resp, _, err := doHTTPReq(req)
	return string(resp), err
}

// RemoveStuff trims spaces, removes new lines and code tag from a string.
func RemoveStuff(s string) string {
	res := strings.Replace(s, "\n", "", -1)
	res = strings.Replace(res, "<code>", "", -1)
	res = strings.Replace(res, "</code>", "", -1)
	res = strings.TrimSpace(res)

	return res
}

// Generate func of custom spaces indentation.
func GenerateListIndentation(spaces int) func() string {
	return func() string {
		return strings.Repeat(" ", spaces)
	}
}

// EscapeSpecChars Escapes special characters
func EscapeSpecChars(s string) string {
	specChar := []string{"\\", "`", "*", "_", "{", "}", "#", "+", "-", ".", "!"}
	res := s

	for _, c := range specChar {
		res = strings.Replace(res, c, "\\"+c, -1)
	}
	return res
}

// ShowHeader shows header befor TOC.
func ShowHeader(w io.Writer) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Table of Contents")
	fmt.Fprintln(w, "=================")
	fmt.Fprintln(w)
}

// ShowFooter shows footer after TOC.
func ShowFooter(w io.Writer) {
	fmt.Fprintln(w, "Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)")
}
