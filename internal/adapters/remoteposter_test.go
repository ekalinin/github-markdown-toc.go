package adapters

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func Test_RemotePoster(t *testing.T) {
	ctx := context.Background()
	testToken := "token-for-test"
	fileName, err := NewFileTemper().CreateTemp(ctx, t.TempDir(), "example.*.txt")
	if err != nil {
		t.Error("Tmp file creation err=", err)
	}
	if err := fileName.Close(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(fileName.Name()); err != nil {
			t.Error("Tmp file deletion err=", err)
		}
	}()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Error("Should be POST")
		}
		tokenGot := r.Header.Get("Authorization")
		tokenWant := "token " + testToken
		if tokenGot != tokenWant {
			t.Error("Auth fail. Want token=", tokenWant, ", got=", tokenGot)
		}

		ctGot := r.Header.Get("Content-Type")
		ctWant := "text/plain;charset=utf-8"
		if ctGot != ctWant {
			t.Error("Content type fail. Want=", ctWant, ", but got=", ctGot)
		}
	}))
	defer srv.Close()

	if _, err := NewRemotePoster().Post(ctx, srv.URL, testToken, fileName.Name()); err != nil {
		t.Error("Should not be err, but got=", err)
	}
}
