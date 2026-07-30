package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
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
			HideHeader: false,
			HideFooter: false,
			Files:      []string{"aaa"},
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
