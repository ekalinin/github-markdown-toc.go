package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/ekalinin/github-markdown-toc.go/v2/internal/version"
)

type TestController struct {
	err  error
	body string
}

func (c TestController) Process(_ context.Context, stdout io.Writer) error {
	if len(c.body) > 0 {
		if _, err := fmt.Fprint(stdout, c.body); err != nil {
			return err
		}

	}
	return c.err
}

func Test_AppRun(t *testing.T) {
	ctl := TestController{}
	app := App{
		cfg: Config{
			Files: []string{"aaa"},
		},
		ctl: ctl,
	}

	var b bytes.Buffer
	if err := app.Run(context.Background(), &b); err != nil {
		t.Error(err)
	}

	want := "\nTable of Contents\n=================\n\n" +
		"Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)\n"
	if got := b.String(); got != want {
		t.Errorf("\nWant=%s\n Got=%s", want, got)
	}
}

func Test_AppRunFail(t *testing.T) {
	errWant := errors.New("process failed")
	ctl := TestController{err: errWant, body: "partial result\n"}
	app := App{
		cfg: Config{},
		ctl: ctl,
	}

	var b bytes.Buffer
	err := app.Run(context.Background(), &b)
	if !errors.Is(err, errWant) {
		t.Errorf("\nWant=%s\n Got=%v", errWant, err)
	}
	if strings.Contains(b.String(), "Created by") {
		t.Errorf("footer was printed after an error: %q", b.String())
	}
	if !strings.Contains(b.String(), "partial result") {
		t.Errorf("successful partial output was lost: %q", b.String())
	}
}

func TestRunWarnsAboutRemoteInputsWithInsert(t *testing.T) {
	var stderr bytes.Buffer
	cfg := Config{
		Files:  []string{"https://github.com/ekalinin/envirius/blob/master/README.md"},
		Insert: InsertConfig{Enabled: true},
		GitHub: GitHubConfig{GHVersion: version.GH_2024_03},
	}
	application, err := New(cfg, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	application.ctl = TestController{}

	if err := application.Run(context.Background(), &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stderr.String(), "is not a local file") {
		t.Errorf("got stderr %q, want the not-a-local-file warning", stderr.String())
	}
}
