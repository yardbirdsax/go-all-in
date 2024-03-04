package json

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
)

const (
	colon   byte = ':'
	comma   byte = ','
	newline byte = '\n'
	quote   byte = '"'
	space   byte = ' '
)

type item struct {
	Foo string
	Bar int64
}

type sample struct {
	Count int
	Items []item
}

type itemWithUnmarshal struct {
	Foo string
	Bar int64
}

func (i *itemWithUnmarshal) UnmarshalJSON(data []byte) (err error) {
	buf := bytes.NewBuffer(data)
	var quote byte = '"'
	for {
		err = seek(buf, quote)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("error scanning for quote: %w", err)
		}
		keyBytes, err := readUntil(buf, quote)
		if err != nil {
			return fmt.Errorf("error reading key: %w", err)
		}
		err = seek(buf, colon)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("error scanning for value quote: %w", err)
		}
		valBytes, err := readValue(buf)
		if err != nil {
			return fmt.Errorf("error reading value: %w", err)
		}
		switch string(keyBytes) {
		case "foo":
			i.Foo = string(valBytes)
		case "bar":
			i.Bar, err = strconv.ParseInt(string(valBytes), 10, 64)
			if err != nil {
				return fmt.Errorf("error decoding Bar value %q: %w", valBytes, err)
			}
		}
	}
	return nil
}

// seek advances a buffer until it hits the given character, or encounters an error.
func seek(buf *bytes.Buffer, char byte) (err error) {
	_, err = readUntil(buf, char)
	return err
}

// readValue reads a buffer to decode a value (i.e., in a JSON key/value pair). It will return
// a string representation of the value.
func readValue(buf *bytes.Buffer) (val []byte, err error) {
	var got byte
	// read until a non-space or quote character is found or EOF
	for {
		got, err = buf.ReadByte()
		if err != nil {
			return val, err
		}
		if got != space && got != quote {
			break
		}
	}
	// read until you get a close quote or a ","
	val = append(val, got)
	for {
		got, err = buf.ReadByte()
		if err != nil {
			return val, err
		}
		if got == quote || got == comma || got == newline {
			break
		}
		val = append(val, got)
	}
	return val, err
}

// readUntil reads a buffer until it hits the given character, or encounters an error. If no error
// is encountered, then it will return the contents of the buffer it read.
func readUntil(buf *bytes.Buffer, char byte) (contents []byte, err error) {
	var got byte
	for {
		got, err = buf.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, io.EOF
			}
			return contents, fmt.Errorf("could not locate starting quote character in buffer: %s", buf.Bytes())
		}
		if got == char {
			break
		}
		contents = append(contents, got)
	}
	return contents, nil
}

type sampleWithUnmarshal struct {
	Count int
	Items []itemWithUnmarshal
}
