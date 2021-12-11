package ghtoc

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const (
	// Version is a current app version
	Version   = "1.2.0"
	userAgent = "github-markdown-toc.go v" + Version
)

// doHTTPReq executes a particullar http request
func doHTTPReq(req *http.Request) ([]byte, string, error) {
	req.Header.Set("User-Agent", userAgent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, "", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, "", err
	}

	if resp.StatusCode == http.StatusForbidden {
		return []byte{}, resp.Header.Get("Content-type"), errors.New(string(body))
	}

	return body, resp.Header.Get("Content-type"), nil
}

// Executes HTTP GET request
func httpGet(urlPath string) ([]byte, string, error) {
	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return []byte{}, "", err
	}
	return doHTTPReq(req)
}

// httpPost executes HTTP POST with file content
func httpPost(urlPath, filePath, token string) (string, error) {
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

// removeStuf trims spaces, removes new lines and code tag from a string
func removeStuf(s string) string {
	res := strings.Replace(s, "\n", "", -1)
	res = strings.Replace(res, "<code>", "", -1)
	res = strings.Replace(res, "</code>", "", -1)
	res = strings.TrimSpace(res)

	return res
}

// generate func of custom spaces indentation
func generateListIndentation(spaces int) func() string {
	return func() string {
		return strings.Repeat(" ", spaces)
	}
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

// ConvertMd2Html Sends Markdown to the github converter
// and returns html
func ConvertMd2Html(localpath string, token string) (string, error) {
	url := "https://api.github.com/markdown/raw"
	return httpPost(url, localpath, token)
}
