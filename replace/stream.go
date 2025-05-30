package replace

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/afero"
)

func ReplaceInFile(path string, old, new string, fs afero.Fs) error {
	fsinfo, err := fs.Stat(path)
	if err != nil {
		return fmt.Errorf("error calling stat on file %q: %w", path, err)
	}
	f, err := fs.OpenFile(path, os.O_RDWR, fsinfo.Mode())
	if err != nil {
		return fmt.Errorf("error opening file %q: %w", path, err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("error reading from file %q: %w", path, err)
	}
	b = bytes.ReplaceAll(b, []byte(old), []byte(new))
	_, err = f.WriteAt(b, 0)
	if err != nil {
		return fmt.Errorf("error writing back to file %q: %w", path, err)
	}
	return nil
}

func ReplaceInFileStreaming(path string, old, new string, fs afero.Fs) error {

	const (
		stateInitial int = iota
		stateParsing
		stateFoundOld
	)

	var (
		currentState int   = stateInitial
		newBytes           = []byte(new)
		newLen             = int64(len(newBytes))
		oldBytes           = []byte(old)
		oldLen             = int64(len(oldBytes))
		readPos      int64 = 0
		writePos     int64 = 0
	)

	fsinfo, err := fs.Stat(path)
	if err != nil {
		return fmt.Errorf("error calling stat on file %q: %w", path, err)
	}
	f, err := fs.OpenFile(path, os.O_RDWR, fsinfo.Mode())
	if err != nil {
		return fmt.Errorf("error opening file %q: %w", path, err)
	}
	defer f.Close()

	readPos = fsinfo.Size() - 1
	writePos = fsinfo.Size() - 1
	oldReadAheadBuf := make([]byte, len(oldBytes))
	singleBuf := make([]byte, 1)
	for {
		if readPos < 0 {
			break
		}
		switch currentState {
		case stateInitial:
			_, err := f.ReadAt(singleBuf, readPos)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return fmt.Errorf("error reading from file at position %d: %w", readPos, err)
			}
			if singleBuf[0] == oldBytes[len(oldBytes)-1] {
				currentState = stateParsing
				continue
			}
			_, err = f.WriteAt(singleBuf, writePos)
			if err != nil {
				return fmt.Errorf("error writing single at position %d: %w", writePos, err)
			}
			readPos--
			writePos--
		case stateParsing:
			if readPos-oldLen+1 < 1 {
				currentState = stateInitial
				_, err = f.WriteAt(singleBuf, writePos)
				if err != nil {
					return fmt.Errorf("error writing single character at position %d: %w", writePos, err)
				}
				readPos--
				writePos--
				continue
			}
			_, err := f.ReadAt(oldReadAheadBuf, readPos-oldLen+1)
			if err != nil {
				return fmt.Errorf("error searching for old at position %d: %w", readPos, err)
			}
			if bytes.Equal(oldBytes, oldReadAheadBuf) {
				currentState = stateFoundOld
				continue
			}
			currentState = stateInitial
			_, err = f.WriteAt(singleBuf, writePos)
			if err != nil {
				return fmt.Errorf("error writing single character at position %d: %w", writePos, err)
			}
			readPos--
			writePos--
		case stateFoundOld:
			_, err := f.WriteAt(newBytes, writePos-newLen+1)
			if err != nil {
				return fmt.Errorf("error writing replacement at position %d: %w", writePos, err)
			}
			readPos -= oldLen
			writePos -= newLen
			currentState = stateInitial
		default:
			return fmt.Errorf("unexpected state %d reached", currentState)
		}
	}
	return nil
}

func readAndShift(buf []byte, readPos int, rdr io.ReaderAt) (b byte, err error) {
	b = buf[0]
	newB := make([]byte, 1)
	_, err = rdr.ReadAt(newB, int64(readPos))
	for i := 1; i < len(buf); i++ {
		buf[i-1] = buf[i]
	}
	buf[len(buf)-1] = newB[0]
	return b, err
}
