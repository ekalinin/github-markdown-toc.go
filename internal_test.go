package ghtoc

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
