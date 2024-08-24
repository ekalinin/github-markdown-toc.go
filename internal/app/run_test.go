package app

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"
)

type TestController struct {
	err  error
	body string
}

func (c TestController) Process(stdout io.Writer) error {
	if c.err != nil {
		return c.err
	}
	if len(c.body) > 0 {
		fmt.Fprint(stdout, c.body)
	}
	return nil
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
	if err := app.Run(&b); err != nil {
		t.Error(err)
	}

	want := "\nTable of Contents\n=================\n\n" +
		"Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc.go)\n"
	if got := b.String(); got != want {
		t.Errorf("\nWant=%s\n Got=%s", want, got)
	}
}

func Test_AppRunFail(t *testing.T) {
	errWant := errors.New("Proccess failed!")
	ctl := TestController{err: errWant}
	app := App{
		cfg: Config{},
		ctl: ctl,
	}

	var b bytes.Buffer
	err := app.Run(&b)
	if err.Error() != errWant.Error() {
		t.Errorf("\nWant=%s\n Got=%s", errWant.Error(), err.Error())
	}
}
