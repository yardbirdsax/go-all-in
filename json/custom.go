package json

import (
	"encoding/json"
	"fmt"
	"os"
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
