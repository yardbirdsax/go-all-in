package mp4

import (
	"fmt"
	"io"
	"strings"
)

// PrintReport writes a formatted inspection report to w.
func PrintReport(w io.Writer, fi *FileInfo) {
	fmt.Fprintf(w, "=== MP4 File Inspector ===\n")
	fmt.Fprintf(w, "Size: %d bytes (%.1f MB)\n", fi.Size, float64(fi.Size)/1024/1024)

	if fi.Ftyp != nil {
		fmt.Fprintf(w, "\n=== Container ===\n")
		fmt.Fprintf(w, "Major Brand: %s\n", fi.Ftyp.MajorBrand)
		fmt.Fprintf(w, "Minor Version: %d\n", fi.Ftyp.MinorVersion)
		if len(fi.Ftyp.CompatibleBrands) > 0 {
			fmt.Fprintf(w, "Compatible: %s\n", strings.Join(fi.Ftyp.CompatibleBrands, ", "))
		}
	}

	fmt.Fprintf(w, "Duration: %.6fs\n", fi.Duration)
	fmt.Fprintf(w, "Overall Bit Rate: %d bps (%.1f Mbps)\n", fi.BitRate, float64(fi.BitRate)/1_000_000)

	fmt.Fprintf(w, "\n=== Box Tree ===\n")
	PrintBoxTree(w, fi.Boxes, 0)

	if fi.Video != nil {
		v := fi.Video
		fmt.Fprintf(w, "\n=== Video Track ===\n")
		fmt.Fprintf(w, "Codec: %s\n", v.Codec)
		fmt.Fprintf(w, "Profile: %s\n", v.Profile)
		fmt.Fprintf(w, "Level: %d\n", v.Level)
		fmt.Fprintf(w, "Resolution: %dx%d\n", v.Width, v.Height)
		if v.FrameRateDen > 0 {
			fps := float64(v.FrameRateNum) / float64(v.FrameRateDen)
			fmt.Fprintf(w, "Frame Rate: %.3f fps (%d/%d)\n", fps, v.FrameRateNum, v.FrameRateDen)
		}
		fmt.Fprintf(w, "Bit Rate: %d bps (%.1f Mbps)\n", v.BitRate, float64(v.BitRate)/1_000_000)
		fmt.Fprintf(w, "Frames: %d\n", v.FrameCount)
		fmt.Fprintf(w, "Keyframes: %d\n", v.KeyFrameCount)
		if len(v.KeyFrames) > 0 {
			fmt.Fprintf(w, "Keyframe Positions: %v\n", v.KeyFrames)
		}
	}

	if fi.Audio != nil {
		a := fi.Audio
		fmt.Fprintf(w, "\n=== Audio Track ===\n")
		fmt.Fprintf(w, "Codec: %s\n", a.Codec)
		fmt.Fprintf(w, "Profile: %s\n", a.Profile)
		fmt.Fprintf(w, "Sample Rate: %d Hz\n", a.SampleRate)
		fmt.Fprintf(w, "Channels: %d\n", a.Channels)
		fmt.Fprintf(w, "Bit Rate: %d bps (%.1f kbps)\n", a.BitRate, float64(a.BitRate)/1000)
		fmt.Fprintf(w, "Frames: %d\n", a.FrameCount)
	}
}

// PrintBoxTree prints an indented tree of boxes.
func PrintBoxTree(w io.Writer, boxes []*Box, indent int) {
	prefix := strings.Repeat("  ", indent)
	for _, b := range boxes {
		fmt.Fprintf(w, "%s[0x%08X] %10d  %s\n", prefix, b.Offset, b.Size, b.Type)
		if len(b.Children) > 0 {
			PrintBoxTree(w, b.Children, indent+1)
		}
	}
}
