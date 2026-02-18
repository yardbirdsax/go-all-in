package main

import (
	"fmt"
	"os"

	"github.com/yardbirdsax/go-all-in/ffprobe/mp4"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file.mp4>\n", os.Args[0])
		os.Exit(1)
	}

	path := os.Args[1]
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error stating file: %v\n", err)
		os.Exit(1)
	}

	boxes, err := mp4.Parse(f, stat.Size())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing MP4: %v\n", err)
		os.Exit(1)
	}

	fi := mp4.Analyze(boxes, stat.Size())
	mp4.PrintReport(os.Stdout, fi)
}
