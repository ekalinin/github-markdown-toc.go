package controller

import (
	"fmt"
	"io"
	"os"
)

func (ctl *Controller) ProcessSTDIN(stdout io.Writer, stding *os.File) error {
	bytes, err := io.ReadAll(stding)
	if err != nil {
		return err
	}

	file, err := os.CreateTemp(os.TempDir(), "ghtoc")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(file.Name()); err != nil {
			_, _ = fmt.Fprintln(stdout, "Error during file delete:", err)
		}
	}()

	err = os.WriteFile(file.Name(), bytes, 0644)
	if err != nil {
		return err
	}

	return ctl.ProcessFiles(stdout, file.Name())
}
