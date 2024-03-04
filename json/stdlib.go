package json

import (
	"encoding/json"
	"fmt"
	"os"
)

type stdlib struct{}

func (s *stdlib) Unmarshal(filename string) (sam *sample, err error) {
	sam = &sample{}
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
