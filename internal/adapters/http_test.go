package adapters

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/version"
)

func testHTTPClient() *http.Client {
	return &http.Client{Timeout: time.Second}
}

func TestHttpGet(t *testing.T) {
	expected := "dummy data"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if ua != version.UserAgent {
			t.Errorf("User-agent should be=%s, got=%s\n", version.UserAgent, ua)
		}

		w.WriteHeader(http.StatusCreated)
		_, err := fmt.Fprint(w, expected)
		if err != nil {
			println(err)
		}
	}))
	defer srv.Close()

	body, _, err := HttpGet(context.Background(), testHTTPClient(), srv.URL)
	got := string(body)

	if err != nil {
		t.Error("Should not be err", err)
	}
	if got != expected {
		t.Error("\nGot :", got, "\nWant:", expected)
	}
}

func TestHttpGetJson(t *testing.T) {
	expected := "dummy data"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if ua != version.UserAgent {
			t.Errorf("User-agent should be=%s, got=%s\n", version.UserAgent, ua)
		}
		want := "application/json"
		ctGot := r.Header.Get("Content-type")
		if ctGot != want {
			t.Errorf("Content-type should be=%s, got=%s\n", want, ctGot)
		}
		acGot := r.Header.Get("Accept")
		if acGot != want {
			t.Errorf("Accept should be=%s, got=%s\n", want, acGot)
		}

		_, err := fmt.Fprint(w, expected)
		if err != nil {
			println(err)
		}
	}))
	defer srv.Close()

	body, _, err := HttpGetJson(context.Background(), testHTTPClient(), srv.URL)
	got := string(body)

	if err != nil {
		t.Error("Should not be err", err)
	}
	if got != expected {
		t.Error("\nGot :", got, "\nWant:", expected)
	}
}

func TestHTTPStatusError(t *testing.T) {
	for _, statusCode := range []int{http.StatusBadRequest, http.StatusInternalServerError} {
		t.Run(http.StatusText(statusCode), func(t *testing.T) {
			const responseBody = "request failed"
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(statusCode)
				_, _ = fmt.Fprint(w, responseBody)
			}))
			defer srv.Close()

			_, _, err := HttpGet(context.Background(), testHTTPClient(), srv.URL)
			var statusErr *HTTPStatusError
			if !errors.As(err, &statusErr) {
				t.Fatalf("got error %v, want HTTPStatusError", err)
			}
			if statusErr.StatusCode != statusCode {
				t.Errorf("got status %d, want %d", statusErr.StatusCode, statusCode)
			}
			if statusErr.Body != responseBody {
				t.Errorf("got body fragment %q, want %q", statusErr.Body, responseBody)
			}
		})
	}
}

func TestHTTPStatusErrorLimitsBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = fmt.Fprint(w, strings.Repeat("x", int(maxErrorBodyBytes)+100))
	}))
	defer srv.Close()

	_, _, err := HttpGet(context.Background(), testHTTPClient(), srv.URL)
	var statusErr *HTTPStatusError
	if !errors.As(err, &statusErr) {
		t.Fatalf("got error %v, want HTTPStatusError", err)
	}
	if got := int64(len(statusErr.Body)); got > maxErrorBodyBytes {
		t.Errorf("got error body length %d, limit is %d", got, maxErrorBodyBytes)
	}
}

func TestHTTPResponseBodyLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", strconv.FormatInt(maxResponseBodyBytes+1, 10))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	_, _, err := HttpGet(context.Background(), testHTTPClient(), srv.URL)
	var tooLargeErr *ResponseBodyTooLargeError
	if !errors.As(err, &tooLargeErr) {
		t.Fatalf("got error %v, want ResponseBodyTooLargeError", err)
	}
}

func createTmp(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "request.md")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestHttpPost(t *testing.T) {
	token := "xxx-token-yyy"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Error("Should be POST")
		}
		tokenPassed := r.Header.Get("Authorization")
		tokenWanted := "token " + token
		if tokenPassed != tokenWanted {
			t.Error("Should pass token", tokenWanted, ", but passed: ", tokenPassed)
		}
	}))
	defer srv.Close()

	fileName := createTmp(t, "#some title")

	_, err := HttpPost(context.Background(), testHTTPClient(), srv.URL, fileName, token)
	if err != nil {
		t.Error("Should not be err", err)
	}
}

func TestHttpPostCancelsInFlightRequest(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		close(started)
		<-release
	}))
	defer func() {
		close(release)
		srv.Close()
	}()

	fileName := createTmp(t, "# some title")

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := HttpPost(ctx, testHTTPClient(), srv.URL, fileName, "")
		done <- err
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("HTTP request did not start")
	}
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("got error %v, want context cancellation", err)
		}
	case <-time.After(time.Second):
		t.Fatal("HTTP request was not canceled")
	}
}

func TestHTTPClientTimeout(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		<-release
	}))
	defer func() {
		close(release)
		srv.Close()
	}()

	client := &http.Client{Timeout: 50 * time.Millisecond}
	_, _, err := HttpGet(context.Background(), client, srv.URL)
	if err == nil {
		t.Fatal("expected an HTTP client timeout")
	}
	var netErr net.Error
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Fatalf("got error %v, want a network timeout", err)
	}
}

func TestHTTPHelpersRejectMalformedURL(t *testing.T) {
	const malformedURL = "://bad-url"
	client := testHTTPClient()
	ctx := context.Background()

	if _, _, err := HttpGet(ctx, client, malformedURL); err == nil {
		t.Error("HttpGet accepted a malformed URL")
	}
	if _, _, err := HttpGetJson(ctx, client, malformedURL); err == nil {
		t.Error("HttpGetJson accepted a malformed URL")
	}
	if _, err := HttpPost(ctx, client, malformedURL, createTmp(t, "body"), ""); err == nil {
		t.Error("HttpPost accepted a malformed URL")
	}
}

func Test_doHTTPReq_issue35(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, "Hello, client")
	}))
	defer srv.Close()

	req, err := http.NewRequest("POST", srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resBody, resHeader, err := doHTTPReq(testHTTPClient(), req)
	if err != nil {
		t.Fatal("doHTTPReq should not be err:", err.Error())
	}
	if string(resBody) != "Hello, client\n" {
		t.Error("response body should be \"Hello, client\", but got:", string(resBody))
	}
	if resHeader != "text/plain; charset=utf-8" {
		t.Error("response header should be \"Hello, client\", but got:", resHeader)
	}
}
