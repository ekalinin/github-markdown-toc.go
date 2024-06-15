package utils

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestHttpGet(t *testing.T) {
	expected := "dummy data"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
