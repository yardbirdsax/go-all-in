package mp4_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yardbirdsax/go-all-in/ffprobe/mp4"
)

const testFile = "/Users/joshuafeierman/Movies/Final/Cat.mp4"

type ffprobeOutput struct {
	Streams []ffprobeStream `json:"streams"`
	Format  ffprobeFormat   `json:"format"`
}

type ffprobeStream struct {
	Index      int    `json:"index"`
	CodecName  string `json:"codec_name"`
	CodecType  string `json:"codec_type"`
	Profile    string `json:"profile"`
	Level      int    `json:"level"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	SampleRate string `json:"sample_rate"`
	Channels   int    `json:"channels"`
	BitRate    string `json:"bit_rate"`
	NbFrames   string `json:"nb_frames"`
}

type ffprobeFormat struct {
	Duration string `json:"duration"`
}

func TestMatchesFFprobe(t *testing.T) {
	// Fail (not skip) if ffprobe is not installed
	ffprobePath, err := exec.LookPath("ffprobe")
	require.NoError(t, err, "ffprobe must be installed to run this test")
	require.NotEmpty(t, ffprobePath)

	// Fail (not skip) if test file is missing
	_, err = os.Stat(testFile)
	require.NoError(t, err, "test file %s must exist", testFile)

	// Run ffprobe
	cmd := exec.Command("ffprobe", "-v", "quiet", "-show_streams", "-show_format", "-print_format", "json", testFile)
	out, err := cmd.Output()
	require.NoError(t, err, "ffprobe failed")

	var ref ffprobeOutput
	require.NoError(t, json.Unmarshal(out, &ref))

	// Find video and audio streams in ffprobe output
	var videoStream, audioStream *ffprobeStream
	for i := range ref.Streams {
		switch ref.Streams[i].CodecType {
		case "video":
			videoStream = &ref.Streams[i]
		case "audio":
			audioStream = &ref.Streams[i]
		}
	}
	require.NotNil(t, videoStream, "ffprobe must report a video stream")
	require.NotNil(t, audioStream, "ffprobe must report an audio stream")

	// Parse with our tool
	f, err := os.Open(testFile)
	require.NoError(t, err)
	defer f.Close()

	stat, err := f.Stat()
	require.NoError(t, err)

	boxes, err := mp4.Parse(f, stat.Size())
	require.NoError(t, err)

	fi := mp4.Analyze(boxes, stat.Size())
	require.NotNil(t, fi.Video, "parser must find video track")
	require.NotNil(t, fi.Audio, "parser must find audio track")

	// Video codec
	assert.Equal(t, videoStream.CodecName, fi.Video.Codec,
		"video codec mismatch")

	// H.264 profile
	assert.Equal(t, videoStream.Profile, fi.Video.Profile,
		"H.264 profile mismatch")

	// H.264 level
	assert.Equal(t, videoStream.Level, fi.Video.Level,
		"H.264 level mismatch")

	// Resolution
	assert.Equal(t, videoStream.Width, fi.Video.Width,
		"video width mismatch")
	assert.Equal(t, videoStream.Height, fi.Video.Height,
		"video height mismatch")

	// Audio codec
	assert.Equal(t, audioStream.CodecName, fi.Audio.Codec,
		"audio codec mismatch")

	// Audio sample rate
	assert.Equal(t, audioStream.SampleRate, fmt.Sprintf("%d", fi.Audio.SampleRate),
		"audio sample rate mismatch")

	// Audio channels
	assert.Equal(t, audioStream.Channels, fi.Audio.Channels,
		"audio channels mismatch")

	// Duration
	assert.Equal(t, ref.Format.Duration, fmt.Sprintf("%.6f", fi.Duration),
		"duration mismatch")

	// Video frame count
	assert.Equal(t, videoStream.NbFrames, fmt.Sprintf("%d", fi.Video.FrameCount),
		"video frame count mismatch")

	// Video bitrate
	assert.Equal(t, videoStream.BitRate, fmt.Sprintf("%d", fi.Video.BitRate),
		"video bitrate mismatch")

	// Audio bitrate
	assert.Equal(t, audioStream.BitRate, fmt.Sprintf("%d", fi.Audio.BitRate),
		"audio bitrate mismatch")
}
