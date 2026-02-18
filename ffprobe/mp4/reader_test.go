package mp4

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSingleBox(t *testing.T) {
	// Build a minimal ftyp box: size(4) + type(4) + major_brand(4) + minor_ver(4) + compat(4) = 20
	data := make([]byte, 20)
	binary.BigEndian.PutUint32(data[0:4], 20)
	copy(data[4:8], "ftyp")
	copy(data[8:12], "mp42")
	binary.BigEndian.PutUint32(data[12:16], 1)
	copy(data[16:20], "isom")

	boxes, err := Parse(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	require.Len(t, boxes, 1)
	assert.Equal(t, "ftyp", boxes[0].Type)
	assert.Equal(t, int64(20), boxes[0].Size)
	assert.Equal(t, int64(0), boxes[0].Offset)
	assert.Equal(t, 12, len(boxes[0].Data)) // payload = 20 - 8 header
}

func TestParseMultipleBoxes(t *testing.T) {
	// Two boxes: ftyp(20) + mdat(16)
	data := make([]byte, 36)
	// ftyp
	binary.BigEndian.PutUint32(data[0:4], 20)
	copy(data[4:8], "ftyp")
	copy(data[8:12], "mp42")
	// mdat
	binary.BigEndian.PutUint32(data[20:24], 16)
	copy(data[24:28], "mdat")
	copy(data[28:36], "testdata")

	boxes, err := Parse(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	require.Len(t, boxes, 2)
	assert.Equal(t, "ftyp", boxes[0].Type)
	assert.Equal(t, "mdat", boxes[1].Type)
	assert.Equal(t, int64(20), boxes[1].Offset)
}

func TestParseContainerBox(t *testing.T) {
	// moov(40) containing mvhd(32)
	// moov header: 8 bytes, mvhd: 32 bytes total = 40
	data := make([]byte, 40)
	binary.BigEndian.PutUint32(data[0:4], 40)
	copy(data[4:8], "moov")
	// child: mvhd
	binary.BigEndian.PutUint32(data[8:12], 32)
	copy(data[12:16], "mvhd")
	// mvhd payload (24 bytes): version=0, flags, creation, modification, timescale=1000, duration=5000
	data[16] = 0 // version
	binary.BigEndian.PutUint32(data[28:32], 1000)  // timescale at offset 12+16=28 in data, offset 12 in payload
	binary.BigEndian.PutUint32(data[32:36], 5000)  // duration

	boxes, err := Parse(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	require.Len(t, boxes, 1)
	assert.Equal(t, "moov", boxes[0].Type)
	require.Len(t, boxes[0].Children, 1)
	assert.Equal(t, "mvhd", boxes[0].Children[0].Type)
}

func TestParseExtendedSize(t *testing.T) {
	// Box with size=1 (extended) and 64-bit size
	data := make([]byte, 24)
	binary.BigEndian.PutUint32(data[0:4], 1)
	copy(data[4:8], "test")
	binary.BigEndian.PutUint64(data[8:16], 24) // extended size includes header
	copy(data[16:24], "payload!")

	boxes, err := Parse(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	require.Len(t, boxes, 1)
	assert.Equal(t, "test", boxes[0].Type)
	assert.Equal(t, int64(24), boxes[0].Size)
	assert.Equal(t, int64(16), boxes[0].HeaderSize)
	assert.Equal(t, []byte("payload!"), boxes[0].Data)
}

func TestFindBox(t *testing.T) {
	boxes := []*Box{
		{Type: "ftyp"},
		{Type: "moov", Children: []*Box{
			{Type: "mvhd"},
			{Type: "trak", Children: []*Box{
				{Type: "tkhd"},
			}},
		}},
	}

	assert.NotNil(t, FindBox(boxes, "ftyp"))
	assert.NotNil(t, FindBox(boxes, "moov", "mvhd"))
	assert.NotNil(t, FindBox(boxes, "moov", "trak", "tkhd"))
	assert.Nil(t, FindBox(boxes, "mdat"))
	assert.Nil(t, FindBox(boxes, "moov", "mdhd"))
}

func TestFindAllBoxes(t *testing.T) {
	boxes := []*Box{
		{Type: "moov", Children: []*Box{
			{Type: "trak", Children: []*Box{{Type: "tkhd"}}},
			{Type: "trak", Children: []*Box{{Type: "tkhd"}}},
		}},
	}

	traks := FindAllBoxes(boxes, "trak")
	assert.Len(t, traks, 2)
	tkhds := FindAllBoxes(boxes, "tkhd")
	assert.Len(t, tkhds, 2)
}

func TestParseFtyp(t *testing.T) {
	data := make([]byte, 12)
	copy(data[0:4], "mp42")
	binary.BigEndian.PutUint32(data[4:8], 1)
	copy(data[8:12], "isom")

	f := parseFtyp(data)
	require.NotNil(t, f)
	assert.Equal(t, "mp42", f.MajorBrand)
	assert.Equal(t, uint32(1), f.MinorVersion)
	assert.Equal(t, []string{"isom"}, f.CompatibleBrands)
}

func TestParseMvhd(t *testing.T) {
	// Version 0: version(1)+flags(3)+create(4)+mod(4)+timescale(4)+duration(4) = 20 bytes minimum
	data := make([]byte, 20)
	data[0] = 0
	binary.BigEndian.PutUint32(data[12:16], 600)
	binary.BigEndian.PutUint32(data[16:20], 7200)

	m := parseMvhd(data)
	require.NotNil(t, m)
	assert.Equal(t, uint32(600), m.Timescale)
	assert.Equal(t, uint64(7200), m.Duration)
}

func TestParseStts(t *testing.T) {
	// version(1)+flags(3)+count(4)+entry(8)
	data := make([]byte, 16)
	binary.BigEndian.PutUint32(data[4:8], 1) // count
	binary.BigEndian.PutUint32(data[8:12], 306)
	binary.BigEndian.PutUint32(data[12:16], 1001)

	s := parseStts(data)
	require.NotNil(t, s)
	require.Len(t, s.Entries, 1)
	assert.Equal(t, uint32(306), s.Entries[0].SampleCount)
	assert.Equal(t, uint32(1001), s.Entries[0].SampleDelta)
}

func TestParseStss(t *testing.T) {
	data := make([]byte, 20)
	binary.BigEndian.PutUint32(data[4:8], 3)
	binary.BigEndian.PutUint32(data[8:12], 1)
	binary.BigEndian.PutUint32(data[12:16], 31)
	binary.BigEndian.PutUint32(data[16:20], 61)

	s := parseStss(data)
	require.NotNil(t, s)
	assert.Equal(t, []uint32{1, 31, 61}, s.SampleNumbers)
}

func TestParseStsz(t *testing.T) {
	// Variable sizes: sample_size=0, count=3, then 3 entries
	data := make([]byte, 24)
	binary.BigEndian.PutUint32(data[4:8], 0) // sample_size
	binary.BigEndian.PutUint32(data[8:12], 3)
	binary.BigEndian.PutUint32(data[12:16], 100)
	binary.BigEndian.PutUint32(data[16:20], 200)
	binary.BigEndian.PutUint32(data[20:24], 300)

	s := parseStsz(data)
	require.NotNil(t, s)
	assert.Equal(t, uint32(3), s.SampleCount)
	assert.Equal(t, []uint32{100, 200, 300}, s.EntrySizes)
	assert.Equal(t, int64(600), s.TotalSize())
}

func TestParseStszUniform(t *testing.T) {
	data := make([]byte, 12)
	binary.BigEndian.PutUint32(data[4:8], 512)
	binary.BigEndian.PutUint32(data[8:12], 100)

	s := parseStsz(data)
	require.NotNil(t, s)
	assert.Equal(t, uint32(100), s.SampleCount)
	assert.Empty(t, s.EntrySizes)
	assert.Equal(t, int64(51200), s.TotalSize())
}

func TestParseAvcC(t *testing.T) {
	// Minimal avcC: configVersion(1)+profile(1)+compat(1)+level(1)+lengthSize(1)+numSPS(1)=6 + no SPS/PPS
	data := []byte{
		1,   // configurationVersion
		100, // profile: High
		0,   // profile compat
		40,  // level 4.0
		0xFF, // lengthSizeMinusOne=3
		0xE0, // numSPS=0
		0,    // numPPS=0
	}

	a := parseAvcC(data)
	require.NotNil(t, a)
	assert.Equal(t, uint8(100), a.ProfileIndication)
	assert.Equal(t, uint8(40), a.LevelIndication)
}

func TestH264ProfileName(t *testing.T) {
	assert.Equal(t, "Baseline", h264ProfileName(66))
	assert.Equal(t, "Main", h264ProfileName(77))
	assert.Equal(t, "High", h264ProfileName(100))
}

func TestAacProfileName(t *testing.T) {
	assert.Equal(t, "LC", aacProfileName(2))
	assert.Equal(t, "Main", aacProfileName(1))
	assert.Equal(t, "SBR", aacProfileName(5))
}

func TestParseEsds(t *testing.T) {
	// Build a minimal ESDS for AAC-LC 48kHz stereo
	// version(1)+flags(3) + ES_Descriptor
	esds := []byte{
		0, 0, 0, 0, // version + flags
		0x03,       // ES_DescrTag
		0x19,       // length = 25
		0x00, 0x01, // ES_ID
		0x00, // flags byte (no deps, no URL, no OCR)
		// DecoderConfigDescriptor
		0x04,       // tag
		0x11,       // length = 17
		0x40,       // objectTypeIndication = Audio ISO 14496-3
		0x15,       // streamType=5 (audio) << 2 | 0x01
		0x00, 0x00, 0x00, // bufferSizeDB
		0x00, 0x01, 0xF4, 0x00, // maxBitrate = 128000
		0x00, 0x01, 0xEA, 0x80, // avgBitrate = 125568
		// DecoderSpecificInfo
		0x05,       // tag
		0x02,       // length = 2
		0x11, 0x90, // AudioSpecificConfig: AAC-LC(2), 48kHz(3), stereo(2)
		// SL descriptor
		0x06, 0x01, 0x02,
	}

	e := parseEsds(esds)
	require.NotNil(t, e)
	assert.Equal(t, uint8(0x40), e.ObjectTypeIndication)
	assert.Equal(t, uint32(128000), e.MaxBitrate)
	assert.Equal(t, uint32(125568), e.AvgBitrate)
	assert.Equal(t, uint8(2), e.AudioObjectType)  // AAC-LC
	assert.Equal(t, uint8(3), e.SamplingFreqIndex) // 48kHz
	assert.Equal(t, uint32(48000), e.SamplingFrequency)
	assert.Equal(t, uint8(2), e.ChannelConfig) // stereo
}
