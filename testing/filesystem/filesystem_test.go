package filesystem

import (
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsert(t *testing.T) {
	existingFilePath := "original.txt"
	insertedLine := "I was appended!"
	mockFs := afero.NewMemMapFs()
	realFs := afero.NewOsFs()
	require.NoError(t, copyToFs(realFs, mockFs, "testdata/original.txt", "original.txt"))

	err := insertLine(existingFilePath, insertedLine, mockFs)

	assert.NoError(t, err, "error was returned when not expected")

	gotBytes, err := afero.ReadFile(mockFs, "original.txt")
	require.NoError(t, err, "error reading got file from mock filesystem")
	wantBytes, err := afero.ReadFile(realFs, "testdata/wanted.txt")
	require.NoError(t, err, "error reading wanted file from real filesystem")
	assert.Equal(t, wantBytes, gotBytes)
}

func copyToFs(sourceFs, destFs afero.Fs, sourcePath string, destPath string) error {
	sourceBytes, err := afero.ReadFile(sourceFs, sourcePath)
	if err != nil {
		return fmt.Errorf("error reading file %q from source filesystem: %w", sourcePath, err)
	}
	err = afero.WriteFile(destFs, destPath, sourceBytes, 0644)
	if err != nil {
		return fmt.Errorf("error writing file %q to dest filesystem: %w", destPath, err)
	}
	return nil
}
