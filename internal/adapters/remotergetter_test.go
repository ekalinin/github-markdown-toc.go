package adapters

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Tpest_RemoteGetterPlain(t *testing.T) {
	expected := "dummy data"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Error("Should be GET")
		}
		_, err := fmt.Fprint(w, expected)
		if err != nil {
			println(err)
		}
	}))
	defer srv.Close()

	getter := NewRemoteGetter(false)
	body, _, err := getter.Get(srv.URL)
	got := string(body)

	if err != nil {
		t.Error("Should not be err", err)
	}
	if got != expected {
		t.Error("\nGot :", got, "\nWant:", expected)
	}
}

func Test_RemoteGetterJson(t *testing.T) {
	expected := "dummy data"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Error("Should be GET")
		}
		ctGot := r.Header.Get("Content-Type")
		ctWant := "application/json"
		if ctGot != ctWant {
			t.Error("Content type fail. Want=", ctWant, ", but got=", ctWant)
		}

		_, err := fmt.Fprint(w, expected)
		if err != nil {
			println(err)
		}
	}))
	defer srv.Close()

	getter := NewRemoteGetter(true)
	body, _, err := getter.Get(srv.URL)
	got := string(body)

	if err != nil {
		t.Error("Should not be err", err)
	}
	if got != expected {
		t.Error("\nGot :", got, "\nWant:", expected)
	}
}
