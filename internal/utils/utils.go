package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/version"
)

const (
	maxResponseBodyBytes int64 = 10 << 20
	maxErrorBodyBytes    int64 = 4 << 10
)

type HTTPStatusError struct {
	StatusCode int
	Body       string
}

func (e *HTTPStatusError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("unexpected HTTP status %d", e.StatusCode)
	}
	return fmt.Sprintf("unexpected HTTP status %d: %s", e.StatusCode, e.Body)
}

type ResponseBodyTooLargeError struct {
	Limit int64
}

func (e *ResponseBodyTooLargeError) Error() string {
	return fmt.Sprintf("HTTP response body exceeds %d bytes", e.Limit)
}

func readLimitedBody(body io.Reader, limit int64) ([]byte, bool, error) {
	data, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		return nil, false, err
	}
	if int64(len(data)) > limit {
		return data[:limit], true, nil
	}
	return data, false, nil
}

// doHTTPReq executes a particular http request
func doHTTPReq(client *http.Client, req *http.Request) ([]byte, string, error) {
	if client == nil {
		return []byte{}, "", errors.New("HTTP client is nil")
	}
	req.Header.Set("User-Agent", version.UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, "", err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	contentType := resp.Header.Get("Content-Type")
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _, err := readLimitedBody(resp.Body, maxErrorBodyBytes)
		if err != nil {
			return []byte{}, contentType, fmt.Errorf("read HTTP error response: %w", err)
		}
		return []byte{}, contentType, &HTTPStatusError{
			StatusCode: resp.StatusCode,
			Body:       strings.TrimSpace(strings.ToValidUTF8(string(body), "\uFFFD")),
		}
	}

	if resp.ContentLength > maxResponseBodyBytes {
		return []byte{}, contentType, &ResponseBodyTooLargeError{Limit: maxResponseBodyBytes}
	}

	body, tooLarge, err := readLimitedBody(resp.Body, maxResponseBodyBytes)
	if err != nil {
		return []byte{}, contentType, err
	}
	if tooLarge {
		return []byte{}, contentType, &ResponseBodyTooLargeError{Limit: maxResponseBodyBytes}
	}

	return body, contentType, nil
}

// HttpGet executes HTTP GET request.
func HttpGet(ctx context.Context, client *http.Client, urlPath string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlPath, nil)
	if err != nil {
		return []byte{}, "", err
	}
	return doHTTPReq(client, req)
}

func HttpGetJson(ctx context.Context, client *http.Client, urlPath string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlPath, nil)
	if err != nil {
		return []byte{}, "", err
	}
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Accept", "application/json")
	return doHTTPReq(client, req)
}

// HttpPost executes HTTP POST with file content.
func HttpPost(ctx context.Context, client *http.Client, url, path, token string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, file)
	if err != nil {
		return "", err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}
	req.ContentLength = fileInfo.Size()

	if token != "" {
		req.Header.Add("Authorization", "token "+token)
	}
	req.Header.Set("Content-Type", "text/plain;charset=utf-8")

	resp, _, err := doHTTPReq(client, req)
	return string(resp), err
}

// RemoveStuff trims spaces, removes new lines and code tag from a string.
func RemoveStuff(s string) string {
	res := strings.ReplaceAll(s, "\n", "")
	res = strings.ReplaceAll(res, "<code>", "")
	res = strings.ReplaceAll(res, "</code>", "")
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
		res = strings.ReplaceAll(res, c, "\\"+c)
	}
	return res
}

// ShowHeader shows header befor TOC.
func ShowHeader(w io.Writer) {
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Table of Contents")
	_, _ = fmt.Fprintln(w, "=================")
	_, _ = fmt.Fprintln(w)
}

// ShowFooter shows footer after TOC.
func ShowFooter(w io.Writer) {
	_, _ = fmt.Fprintln(w, "Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)")
}
