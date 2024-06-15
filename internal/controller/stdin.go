package controller

import (
	"io"
	"os"
)

func (ctl *Controller) ProcessSTDIN(stding *os.File) error {
	bytes, err := io.ReadAll(stding)
	if err != nil {
		return err
	}

	file, err := os.CreateTemp(os.TempDir(), "ghtoc")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	err = os.WriteFile(file.Name(), bytes, 0644)
	if err != nil {
		return err
	}

	return ctl.ProcessFiles(file.Name())
}
