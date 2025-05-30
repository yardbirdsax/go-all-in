package replace

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplaceInFile(t *testing.T) {
	testCases := []struct {
		name        string
		inFile      string
		original    string
		replacement string
		wantedFile  string
	}{
		{
			name:        "simple",
			inFile:      "simple.txt",
			original:    "world",
			replacement: "universe",
			wantedFile:  "simple_wanted.txt",
		},
	}

	for i := range testCases {
		tc := &testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			fullInPath := filepath.Join("testdata", tc.inFile)
			fullTempPath := filepath.Join(os.TempDir(), tc.name+"_"+tc.inFile)
			fullWantedPath := filepath.Join("testdata", tc.wantedFile)
			copyFile(t, fullInPath, fullTempPath)
			t.Cleanup(func() {
				os.Remove(fullTempPath)
			})

			err := ReplaceInFile(fullTempPath, tc.original, tc.replacement, afero.NewOsFs())

			assert.NoError(t, err, "error returned when not wanted")
			assertFilesEqual(t, fullWantedPath, fullTempPath, "got and wanted files differ")
		})
	}
}

func TestReplaceInFileStreaming(t *testing.T) {
	testCases := []struct {
		name        string
		inFile      string
		original    string
		replacement string
		wantedFile  string
	}{
		{
			name:        "simple",
			inFile:      "simple.txt",
			original:    "universe",
			replacement: "world",
			wantedFile:  "simple_wanted.txt",
		},
	}

	for i := range testCases {
		tc := &testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			fullInPath := filepath.Join("testdata", tc.inFile)
			fullTempPath := filepath.Join(os.TempDir(), tc.name+"_"+tc.inFile)
			fullWantedPath := filepath.Join("testdata", tc.wantedFile)
			copyFile(t, fullInPath, fullTempPath)
			t.Cleanup(func() {
				os.Remove(fullTempPath)
			})

			err := ReplaceInFileStreaming(fullTempPath, tc.original, tc.replacement, afero.NewOsFs())

			assert.NoError(t, err, "error returned when not wanted")
			assertFilesEqual(t, fullWantedPath, fullTempPath, "got and wanted files differ")
		})
	}
}

func assertFilesEqual(t *testing.T, wanted, got, message string) {
	wantedF, err := os.Open(wanted)
	require.NoError(t, err, "error opening wanted file")
	defer wantedF.Close()
	gotF, err := os.Open(got)
	require.NoError(t, err, "error opening got file")
	defer gotF.Close()

	line := 0
	wantedBuf := bufio.NewScanner(wantedF)
	gotBuf := bufio.NewScanner(gotF)
	diffB := &strings.Builder{}
	for wantedBuf.Scan() {
		line++
		gotBuf.Scan()
		if diff := cmp.Diff(wantedBuf.Text(), gotBuf.Text()); diff != "" {
			diffB.WriteString(fmt.Sprintf("line %d:\n", line))
			diffB.WriteString(indent(diff, 2))
		}
	}
	if diffB.Len() > 0 {
		t.Errorf("%s\n%s", message, diffB.String())
	}
}

type testinterface interface {
	Errorf(format string, args ...any)
	FailNow()
	Fatalf(format string, args ...any)
}

func copyFile(t testinterface, in, out string) {
	inF, err := os.Open(in)
	require.NoError(t, err, "error opening source file")
	defer inF.Close()
	outF, err := os.OpenFile(out, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0700)
	require.NoError(t, err, "error opening destination file")
	defer outF.Close()
	_, err = io.Copy(outF, inF)
	require.NoError(t, err, "error copying from source to destination")
}

func indent(s string, count int) string {
	b := &strings.Builder{}
	for _, sp := range strings.Split(s, "\n") {
		b.WriteString(fmt.Sprintf("%s%s\n", strings.Repeat(" ", count), sp))
	}
	return b.String()
}

func TestReadAndShift(t *testing.T) {
	testCases := []struct {
		in      string
		inBuf   string
		readPos int
		wantB   byte
		wantBuf string
	}{
		{
			in:      "hello world",
			inBuf:   "hello",
			readPos: 5,
			wantB:   'h',
			wantBuf: "ello ",
		},
		{
			in:      "hello world",
			inBuf:   "ello ",
			readPos: 6,
			wantB:   'e',
			wantBuf: "llo w",
		},
		{
			in:      "hello world",
			inBuf:   "worl",
			readPos: 10,
			wantB:   'w',
			wantBuf: "orld",
		},
		{
			in:      "hello world",
			inBuf:   "d",
			readPos: 12,
			wantB:   'd',
			wantBuf: "\x00",
		},
	}

	for i := range testCases {
		tc := &testCases[i]
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			inBufB := []byte(tc.inBuf)
			gotBuf := make([]byte, len(inBufB))
			copy(gotBuf, inBufB)
			rdr := bytes.NewReader([]byte(tc.in))

			gotB, err := readAndShift(gotBuf, tc.readPos, rdr)

			assert.NoError(t, err, "got error when not wanted")
			assert.Equal(t, tc.wantBuf, string(gotBuf), "got and wanted buffer differ")
			assert.Equal(t, tc.wantB, gotB, "got and wanted byte differ")
		})
	}
}

func BenchmarkReplaceInFile(b *testing.B) {
	filename := "testdata/simple.txt"
	filebytes, err := os.ReadFile(filename)
	require.NoError(b, err, "error reading test file")
	memFs := afero.NewOsFs()
	b.Run("k=ReplaceInFile", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			// Setup logic (not included in timing)
			tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("benchmark_%d.txt", i))
			afero.WriteFile(memFs, tempFile, filebytes, 0700)
			b.StartTimer()

			// Code to benchmark
			err := ReplaceInFile(tempFile, "universe", "world", memFs)
			if err != nil {
				b.Fatalf("unexpected error: %v", err)
			}
		}
	})

	b.Run("k=ReplaceInFileStreaming", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			// Setup logic (not included in timing)
			tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("benchmark_%d.txt", i))
			afero.WriteFile(memFs, tempFile, filebytes, 0700)
			b.StartTimer()

			// Code to benchmark
			err := ReplaceInFileStreaming(tempFile, "universe", "world", memFs)
			if err != nil {
				b.Fatalf("unexpected error: %v", err)
			}
		}
	})
}
