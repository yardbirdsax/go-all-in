package filesystem

import (
	"fmt"
	"os"

	"github.com/spf13/afero"
)

func insertLine(path string, lineToInsert string, fs afero.Fs) (err error) {
	f, err := fs.OpenFile(path, os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("error opening file %q: %w", path, err)
	}
	defer f.Close()
	_, err = f.WriteString("\n")
	if err != nil {
		return fmt.Errorf("error writing to file %q: %w", path, err)
	}
	_, err = f.WriteString(lineToInsert)
	if err != nil {
		return fmt.Errorf("error writing to file %q: %w", path, err)
	}
	return nil
}
