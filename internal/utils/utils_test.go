package utils

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ekalinin/github-markdown-toc.go/internal/version"
)

func TestHttpGet(t *testing.T) {
	expected := "dummy data"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		if ua != version.UserAgent {
			t.Errorf("User-agent should be=%s, got=%s\n", version.UserAgent, ua)
		}

		_, err := fmt.Fprint(w, expected)
		if err != nil {
			println(err)
		}
	}))
	defer srv.Close()

	body, _, err := HttpGet(srv.URL)
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

	body, _, err := HttpGetJson(srv.URL)
	got := string(body)

	if err != nil {
		t.Error("Should not be err", err)
	}
	if got != expected {
		t.Error("\nGot :", got, "\nWant:", expected)
	}
}

func TestHttpGetForbidden(t *testing.T) {
	txt := "please, do not try"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, err := fmt.Fprint(w, txt)
		if err != nil {
			println(err)
		}
	}))
	defer srv.Close()

	_, _, err := HttpGet(srv.URL)
	if err == nil {
		t.Error("Should not not be nil")
	}
}

func createTmp(content string) (string, error) {
	tmpFile, err := os.CreateTemp("", "example.*.txt")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		if err := tmpFile.Close(); err != nil {
			return "", err
		}
		log.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		log.Fatal(err)
	}

	return tmpFile.Name(), nil
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

	fileName, err := createTmp("#some title")
	if err != nil {
		t.Error("Should not be err", err)
	}
	defer os.Remove(fileName)

	_, err = HttpPost(srv.URL, fileName, token)
	if err != nil {
		t.Error("Should not be err", err)
	}
}

// Cover the changes of ioutil.ReadAll to io.ReadAll in doHTTPReq.
func Test_doHTTPReq_issue35(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer srv.Close()

	dummyURL := srv.URL

	req, err := http.NewRequest("POST", dummyURL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resBody, resHeader, err := doHTTPReq(req)

	// Require no error
	if err != nil {
		t.Fatal("doHTTPReq should not be err:", err.Error())
	}

	// Assert response body
	if string(resBody) != "Hello, client\n" {
		t.Error("response body should be \"Hello, client\", but got:", string(resBody))
	}
	// Assert response header
	if resHeader != "text/plain; charset=utf-8" {
		t.Error("response header should be \"Hello, client\", but got:", resHeader)
	}
}

func Test_RemoveStuff(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"All", "\n\nsome<code> code </code> here\n", "some code  here"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RemoveStuff(tt.in)
			if got != tt.want {
				t.Errorf("Got=%s, want=%s\n", got, tt.want)
			}
		})
	}
}

func Test_GenerateListIndentations(t *testing.T) {
	f := GenerateListIndentation(2)
	if got := f(); got != "  " {
		t.Errorf("Got='%s', want='  '", got)
	}
}

func Test_EscapeSpecChars(t *testing.T) {
	in := `abc\*_{}`
	want := "abc\\\\\\*\\_\\{\\}"
	got := EscapeSpecChars(in)
	if got != want {
		t.Errorf("Got=%s, want=%s", got, want)
	}
}

func Test_ShowHeaderFooter(t *testing.T) {
	var b bytes.Buffer

	ShowHeader(&b)
	want := "\nTable of Contents\n=================\n\n"
	if got := b.String(); got != want {
		t.Errorf("\nWant=%s\n Got=%s", want, got)
	}

	b.Reset()
	ShowFooter(&b)
	want = "Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)\n"
	if got := b.String(); got != want {
		t.Errorf("\nWant=%s\n Got=%s", want, got)
	}
}
