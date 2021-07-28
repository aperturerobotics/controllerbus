package hot_compiler

import (
	"io"
	"os"
)

// copyFileFromTo copies a file from a path to another path
func copyFileFromTo(src, dest string) error {
	ofile, oerr := os.Open(src)
	if oerr != nil {
		return oerr
	}
	defer ofile.Close()

	outFileFd, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer outFileFd.Close()
	defer outFileFd.Sync()

	_, err = io.Copy(outFileFd, ofile)
	return err
}
