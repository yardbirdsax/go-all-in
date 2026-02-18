package mp4

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Box represents an MP4 box (atom) in the file.
type Box struct {
	Type       string
	Size       int64
	Offset     int64 // absolute file offset of box start
	HeaderSize int64 // 8 or 16 bytes
	Children   []*Box
	Data       []byte // raw payload for leaf boxes; header bytes for sample entries
}

// containerTypes holds box types whose payload is entirely child boxes.
var containerTypes = map[string]bool{
	"moov": true, "trak": true, "mdia": true, "minf": true,
	"stbl": true, "udta": true, "dinf": true, "edts": true,
	"mvex": true, "moof": true, "traf": true, "sinf": true,
	"schi": true, "ilst": true,
}

// fullContainerTypes are containers with version(1)+flags(3) before children.
var fullContainerTypes = map[string]bool{
	"meta": true,
}

// sampleEntrySkip defines the header byte count after the standard box header
// for sample entry types; children follow after these bytes.
var sampleEntrySkip = map[string]int{
	"avc1": 78,
	"mp4a": 28,
}

// Parse reads all boxes from r (which has the given total size) and returns
// the top-level box tree.
func Parse(r io.ReadSeeker, size int64) ([]*Box, error) {
	return parseChildren(r, 0, size)
}

func parseChildren(r io.ReadSeeker, start, end int64) ([]*Box, error) {
	var boxes []*Box
	pos := start
	for pos < end {
		if end-pos < 8 {
			break
		}
		if _, err := r.Seek(pos, io.SeekStart); err != nil {
			return nil, err
		}
		box, err := readBoxHeader(r, pos, end)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		payloadStart := box.Offset + box.HeaderSize
		payloadEnd := box.Offset + box.Size

		switch {
		case containerTypes[box.Type]:
			children, err := parseChildren(r, payloadStart, payloadEnd)
			if err != nil {
				return nil, fmt.Errorf("parsing children of %s: %w", box.Type, err)
			}
			box.Children = children

		case fullContainerTypes[box.Type]:
			// skip version(1) + flags(3) then parse children
			children, err := parseChildren(r, payloadStart+4, payloadEnd)
			if err != nil {
				return nil, fmt.Errorf("parsing children of %s: %w", box.Type, err)
			}
			box.Children = children

		case box.Type == "stsd":
			// full box (version+flags=4) + entry_count(4) then child boxes
			children, err := parseChildren(r, payloadStart+8, payloadEnd)
			if err != nil {
				return nil, fmt.Errorf("parsing children of stsd: %w", err)
			}
			box.Children = children

		default:
			if skip, ok := sampleEntrySkip[box.Type]; ok {
				// Read sample-entry header into Data, then parse child boxes
				if _, err := r.Seek(payloadStart, io.SeekStart); err != nil {
					return nil, err
				}
				box.Data = make([]byte, skip)
				if _, err := io.ReadFull(r, box.Data); err != nil {
					return nil, fmt.Errorf("reading %s header: %w", box.Type, err)
				}
				children, err := parseChildren(r, payloadStart+int64(skip), payloadEnd)
				if err != nil {
					return nil, fmt.Errorf("parsing children of %s: %w", box.Type, err)
				}
				box.Children = children
			} else {
				// Leaf box — read raw payload
				dataSize := payloadEnd - payloadStart
				if dataSize > 0 && dataSize <= 10*1024*1024 {
					if _, err := r.Seek(payloadStart, io.SeekStart); err != nil {
						return nil, err
					}
					box.Data = make([]byte, dataSize)
					if _, err := io.ReadFull(r, box.Data); err != nil {
						return nil, fmt.Errorf("reading %s data: %w", box.Type, err)
					}
				}
			}
		}

		boxes = append(boxes, box)
		pos = box.Offset + box.Size
	}
	return boxes, nil
}

func readBoxHeader(r io.ReadSeeker, pos, end int64) (*Box, error) {
	var header [8]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, err
	}
	size := int64(binary.BigEndian.Uint32(header[0:4]))
	boxType := string(header[4:8])

	box := &Box{
		Type:       boxType,
		Offset:     pos,
		HeaderSize: 8,
	}

	if size == 1 {
		var ext [8]byte
		if _, err := io.ReadFull(r, ext[:]); err != nil {
			return nil, err
		}
		size = int64(binary.BigEndian.Uint64(ext[:]))
		box.HeaderSize = 16
	} else if size == 0 {
		size = end - pos
	}

	box.Size = size
	return box, nil
}

// FindBox searches the box tree for a box matching the given type path.
// e.g. FindBox(boxes, "moov", "mvhd") finds mvhd inside moov.
func FindBox(boxes []*Box, path ...string) *Box {
	if len(path) == 0 {
		return nil
	}
	for _, b := range boxes {
		if b.Type == path[0] {
			if len(path) == 1 {
				return b
			}
			return FindBox(b.Children, path[1:]...)
		}
	}
	return nil
}

// FindAllBoxes returns every box of the given type at any depth.
func FindAllBoxes(boxes []*Box, boxType string) []*Box {
	var result []*Box
	for _, b := range boxes {
		if b.Type == boxType {
			result = append(result, b)
		}
		result = append(result, FindAllBoxes(b.Children, boxType)...)
	}
	return result
}
