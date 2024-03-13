package json

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

type custom struct{}

func (c *custom) Unmarshal(filename string) (sam *sampleWithUnmarshal, err error) {
	sam = &sampleWithUnmarshal{}
	f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening file %q: %w", filename, err)
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(sam)
	if err != nil {
		return nil, fmt.Errorf("error decoding data: %w", err)
	}
	return sam, err
}

func (c *custom) UnmarshalV2(filename string) (sam *sample, err error) {
	sam = &sample{}
	f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening file %q: %w", filename, err)
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	err = unmarshalSample(buf, sam)
	return sam, err
}

func unmarshalSample(buf io.ByteReader, sam *sample) (err error) {
	// seek to key
	err = seek(buf, quote)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return err
		}
		return err
	}
	key, err := readValue(buf)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return io.ErrUnexpectedEOF
		}
		return err
	}
	switch string(key) {
	case "count":
		err = seek(buf, colon)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return io.ErrUnexpectedEOF
			}
			return err
		}
		countBytes, err := readValue(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return io.ErrUnexpectedEOF
			}
			return err
		}
		sam.Count, err = strconv.ParseInt(string(countBytes), 10, 64)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return io.ErrUnexpectedEOF
			}
			return err
		}
	case "items":
		err = unmarshalItems(buf, sam)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return io.ErrUnexpectedEOF
			}
			return err
		}
	}
	return nil
}

func unmarshalItems(buf io.ByteReader, sam *sample) (err error) {
	err = seek(buf, leftBracket)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return io.ErrUnexpectedEOF
		}
	}
	for {
		err = seek(buf, leftBrace)
		if err != nil {
			// empty array
			if errors.Is(err, io.EOF) {
				return nil
			}
		}
		// read until the closing brace
		itemBuf := bytes.NewBuffer([]byte{})
		
		for {
			i := &item{}
			err = seek(buf, quote)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return io.ErrUnexpectedEOF
				}
				return err
			}
			keyBytes, err := readValue(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return io.ErrUnexpectedEOF
				}
			}
			switch string(keyBytes) {
			case "foo":
				fooBytes, err := readValue(buf)
				if err != nil {
					if errors.Is(err, io.EOF) {
						return io.ErrUnexpectedEOF
					}
					return err
				}
				i.Foo = string(fooBytes)
			}
		}
	}
	return nil
}
