package entity

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func Test_TocPrint(t *testing.T) {
	tests := []struct {
		name string
		toc  *Toc
		want string
	}{
		{"Print", &Toc{"hello", "there"}, "hello\nthere\n\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b bytes.Buffer
			if err := tt.toc.Print(&b); err != nil {
				t.Errorf("failed print, err=%v", err)
			}
			if got := b.String(); got != tt.want {
				t.Errorf("Got=%s, want=%s", got, tt.want)
			}
		})
	}
}

func Test_TocAt(t *testing.T) {
	toc := Toc{"hello", "there"}
	got := toc.At(1)
	if got != "there" {
		t.Errorf("got: %s, want: %s\n", got, "there")
	}
}

type TestPrinter struct {
	n   int
	err string
}

func (p TestPrinter) Fprintln(w io.Writer, a ...any) (n int, err error) {
	if p.err != "" {
		return 0, errors.New(p.err)
	}
	return p.n, nil
}

func Test_TocCustomPrintFail(t *testing.T) {
	toc := Toc{"hello", "there"}
	printer := TestPrinter{0, "failed"}

	var b bytes.Buffer
	got := toc.CustomPrint(&b, printer)

	if got == nil {
		t.Errorf("should fail first print")
	}

	toc = Toc{}
	got = toc.CustomPrint(&b, printer)

	if got == nil {
		t.Errorf("should fail last print")
	}
}
