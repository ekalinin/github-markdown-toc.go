package ghtoc

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func Test_httpGet(t *testing.T) {
	expected := "dummy data"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprint(w, expected)
		if err != nil {
			println(err)
		}
	}))
	defer srv.Close()

	body, _, err := httpGet(srv.URL)
	got := string(body)

	if err != nil {
		t.Error("Should not be err", err)
	}
	if got != expected {
		t.Error("\nGot :", got, "\nWant:", expected)
	}
}

func Test_httpGetForbidden(t *testing.T) {
	txt := "please, do not try"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, err := fmt.Fprint(w, txt)
		if err != nil {
			println(err)
		}
	}))
	defer srv.Close()

	_, _, err := httpGet(srv.URL)
	if err == nil {
		t.Error("Should not not be nil")
	}
}

func createTmp(content string) (string, error) {
	tmpFile, err := ioutil.TempFile("", "example.*.txt")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		tmpFile.Close()
		log.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		log.Fatal(err)
	}

	return tmpFile.Name(), nil
}

func Test_httpPost(t *testing.T) {
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

	_, err = httpPost(srv.URL, fileName, token)
	if err != nil {
		t.Error("Should not be err", err)
	}
}
