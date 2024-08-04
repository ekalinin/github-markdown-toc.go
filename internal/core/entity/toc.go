package entity

import (
	"fmt"
	"io"
)

type Toc []string

func (toc *Toc) Print(w io.Writer) error {
	for _, tocItem := range *toc {
		if _, err := fmt.Fprintln(w, tocItem); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	return nil
}

func (toc Toc) At(idx int) string {
	ss := []string(toc)
	return ss[idx]
}
