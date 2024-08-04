package entity

import (
	"fmt"
	"io"
)

type TocPrinter interface {
	Fprintln(w io.Writer, a ...any) (n int, err error)
}

type TocPrinterDefault struct {
}

func (p TocPrinterDefault) Fprintln(w io.Writer, a ...any) (n int, err error) {
	return fmt.Fprintln(w, a...)
}

type Toc []string

func (toc *Toc) Print(w io.Writer) error {
	printer := TocPrinterDefault{}
	return toc.CustomPrint(w, printer)
}

func (toc *Toc) CustomPrint(w io.Writer, p TocPrinter) error {
	for _, tocItem := range *toc {
		if _, err := p.Fprintln(w, tocItem); err != nil {
			return err
		}
	}
	if _, err := p.Fprintln(w); err != nil {
		return err
	}
	return nil
}

func (toc Toc) At(idx int) string {
	ss := []string(toc)
	return ss[idx]
}
