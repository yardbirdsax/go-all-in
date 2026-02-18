package mp4

import (
	"encoding/binary"
)

// --- Raw parsed box types ---

type FtypBox struct {
	MajorBrand       string
	MinorVersion     uint32
	CompatibleBrands []string
}

type MvhdBox struct {
	Version   uint8
	Timescale uint32
	Duration  uint64
}

type TkhdBox struct {
	Version  uint8
	TrackID  uint32
	Duration uint64
	Width    uint32 // fixed-point 16.16
	Height   uint32 // fixed-point 16.16
}

type MdhdBox struct {
	Version   uint8
	Timescale uint32
	Duration  uint64
}

type HdlrBox struct {
	HandlerType string
	Name        string
}

type SttsEntry struct {
	SampleCount uint32
	SampleDelta uint32
}

type SttsBox struct {
	Entries []SttsEntry
}

type StssBox struct {
	SampleNumbers []uint32
}

type StszBox struct {
	SampleSize  uint32
	SampleCount uint32
	EntrySizes  []uint32
}

// TotalSize returns the sum of all sample sizes.
func (s *StszBox) TotalSize() int64 {
	if s.SampleSize > 0 {
		return int64(s.SampleSize) * int64(s.SampleCount)
	}
	var total int64
	for _, sz := range s.EntrySizes {
		total += int64(sz)
	}
	return total
}

type AvcCBox struct {
	ProfileIndication    uint8
	ProfileCompatibility uint8
	LevelIndication      uint8
	LengthSizeMinusOne   uint8
	SPS                  [][]byte
	PPS                  [][]byte
}

type EsdsBox struct {
	ObjectTypeIndication uint8
	StreamType           uint8
	MaxBitrate           uint32
	AvgBitrate           uint32
	AudioObjectType      uint8
	SamplingFreqIndex    uint8
	SamplingFrequency    uint32
	ChannelConfig        uint8
}

type Mp4aBox struct {
	ChannelCount uint16
	SampleSize   uint16
	SampleRate   uint32 // actual Hz (not fixed-point)
}

type Avc1Box struct {
	Width  uint16
	Height uint16
}

// --- Computed track info ---

type VideoTrack struct {
	Codec         string // "h264"
	Profile       string // "High"
	Level         int    // 40
	Width         int
	Height        int
	FrameRateNum  int // e.g. 24000
	FrameRateDen  int // e.g. 1001
	BitRate       int64
	FrameCount    int
	KeyFrameCount int
	KeyFrames     []uint32
}

type AudioTrack struct {
	Codec      string // "aac"
	Profile    string // "LC"
	SampleRate int
	Channels   int
	BitRate    int64
	FrameCount int
}

// FileInfo holds all parsed + computed info for an MP4 file.
type FileInfo struct {
	Size     int64
	Duration float64 // seconds
	BitRate  int64
	Ftyp     *FtypBox
	Mvhd     *MvhdBox
	Boxes    []*Box
	Video    *VideoTrack
	Audio    *AudioTrack
}

// Analyze walks the parsed box tree and produces a FileInfo.
func Analyze(boxes []*Box, fileSize int64) *FileInfo {
	fi := &FileInfo{
		Size:  fileSize,
		Boxes: boxes,
	}

	if b := FindBox(boxes, "ftyp"); b != nil && b.Data != nil {
		fi.Ftyp = parseFtyp(b.Data)
	}
	if b := FindBox(boxes, "moov", "mvhd"); b != nil && b.Data != nil {
		fi.Mvhd = parseMvhd(b.Data)
	}

	traks := FindAllBoxes(boxes, "trak")
	for _, trak := range traks {
		hdlrBox := FindBox(trak.Children, "mdia", "hdlr")
		if hdlrBox == nil || hdlrBox.Data == nil {
			continue
		}
		hdlr := parseHdlr(hdlrBox.Data)
		if hdlr == nil {
			continue
		}

		// Parse common track boxes
		var mdhd *MdhdBox
		var tkhd *TkhdBox
		var stts *SttsBox
		var stss *StssBox
		var stsz *StszBox

		if b := FindBox(trak.Children, "tkhd"); b != nil && b.Data != nil {
			tkhd = parseTkhd(b.Data)
		}
		if b := FindBox(trak.Children, "mdia", "mdhd"); b != nil && b.Data != nil {
			mdhd = parseMdhd(b.Data)
		}
		if b := FindBox(trak.Children, "mdia", "minf", "stbl", "stts"); b != nil && b.Data != nil {
			stts = parseStts(b.Data)
		}
		if b := FindBox(trak.Children, "mdia", "minf", "stbl", "stss"); b != nil && b.Data != nil {
			stss = parseStss(b.Data)
		}
		if b := FindBox(trak.Children, "mdia", "minf", "stbl", "stsz"); b != nil && b.Data != nil {
			stsz = parseStsz(b.Data)
		}

		switch hdlr.HandlerType {
		case "vide":
			fi.Video = buildVideoTrack(trak, tkhd, mdhd, stts, stss, stsz)
		case "soun":
			fi.Audio = buildAudioTrack(trak, tkhd, mdhd, stts, stsz)
		}
	}

	// Compute file-level duration + bitrate
	if fi.Mvhd != nil && fi.Mvhd.Timescale > 0 {
		fi.Duration = float64(fi.Mvhd.Duration) / float64(fi.Mvhd.Timescale)
	}
	if fi.Duration > 0 {
		fi.BitRate = int64(float64(fi.Size*8) / fi.Duration)
	}

	return fi
}

func buildVideoTrack(trak *Box, tkhd *TkhdBox, mdhd *MdhdBox, stts *SttsBox, stss *StssBox, stsz *StszBox) *VideoTrack {
	vt := &VideoTrack{Codec: "h264"}

	// Resolution from avc1 header
	if avc1Boxes := FindAllBoxes(trak.Children, "avc1"); len(avc1Boxes) > 0 {
		if a := parseAvc1(avc1Boxes[0]); a != nil {
			vt.Width = int(a.Width)
			vt.Height = int(a.Height)
		}
	}
	// Fallback to tkhd
	if vt.Width == 0 && tkhd != nil {
		vt.Width = int(tkhd.Width >> 16)
		vt.Height = int(tkhd.Height >> 16)
	}

	// Profile + level from avcC
	if avcCBoxes := FindAllBoxes(trak.Children, "avcC"); len(avcCBoxes) > 0 && avcCBoxes[0].Data != nil {
		if ac := parseAvcC(avcCBoxes[0].Data); ac != nil {
			vt.Profile = h264ProfileName(ac.ProfileIndication)
			vt.Level = int(ac.LevelIndication)
		}
	}

	// Frame rate from stts + mdhd
	if stts != nil && mdhd != nil && mdhd.Timescale > 0 {
		var totalSamples, totalTicks int64
		for _, e := range stts.Entries {
			totalSamples += int64(e.SampleCount)
			totalTicks += int64(e.SampleCount) * int64(e.SampleDelta)
		}
		if totalTicks > 0 {
			num := totalSamples * int64(mdhd.Timescale)
			den := totalTicks
			g := gcd(num, den)
			vt.FrameRateNum = int(num / g)
			vt.FrameRateDen = int(den / g)
		}
	}

	// Frame count + bitrate from stsz + mdhd
	if stsz != nil {
		vt.FrameCount = int(stsz.SampleCount)
		if mdhd != nil && mdhd.Duration > 0 {
			vt.BitRate = stsz.TotalSize() * 8 * int64(mdhd.Timescale) / int64(mdhd.Duration)
		}
	}

	// Keyframes from stss
	if stss != nil {
		vt.KeyFrameCount = len(stss.SampleNumbers)
		vt.KeyFrames = stss.SampleNumbers
	}

	return vt
}

func buildAudioTrack(trak *Box, _ *TkhdBox, mdhd *MdhdBox, _ *SttsBox, stsz *StszBox) *AudioTrack {
	at := &AudioTrack{Codec: "aac"}

	// Sample rate + channels from mp4a header
	if mp4aBoxes := FindAllBoxes(trak.Children, "mp4a"); len(mp4aBoxes) > 0 {
		if m := parseMp4a(mp4aBoxes[0]); m != nil {
			at.SampleRate = int(m.SampleRate)
			at.Channels = int(m.ChannelCount)
		}
	}

	// Profile from esds AudioSpecificConfig
	var esds *EsdsBox
	if esdsBoxes := FindAllBoxes(trak.Children, "esds"); len(esdsBoxes) > 0 && esdsBoxes[0].Data != nil {
		esds = parseEsds(esdsBoxes[0].Data)
		if esds != nil {
			at.Profile = aacProfileName(esds.AudioObjectType)
			// Override sample rate / channels from AudioSpecificConfig if available
			if esds.SamplingFrequency > 0 {
				at.SampleRate = int(esds.SamplingFrequency)
			}
			if esds.ChannelConfig > 0 {
				at.Channels = int(esds.ChannelConfig)
			}
		}
	}

	// Frame count from stsz
	if stsz != nil {
		at.FrameCount = int(stsz.SampleCount)
	}

	// Bitrate from stsz + mdhd
	if stsz != nil && mdhd != nil && mdhd.Duration > 0 {
		at.BitRate = stsz.TotalSize() * 8 * int64(mdhd.Timescale) / int64(mdhd.Duration)
	}

	return at
}

// --- Individual box parsers ---

func parseFtyp(data []byte) *FtypBox {
	if len(data) < 8 {
		return nil
	}
	f := &FtypBox{
		MajorBrand:   string(data[0:4]),
		MinorVersion: binary.BigEndian.Uint32(data[4:8]),
	}
	for i := 8; i+4 <= len(data); i += 4 {
		f.CompatibleBrands = append(f.CompatibleBrands, string(data[i:i+4]))
	}
	return f
}

func parseMvhd(data []byte) *MvhdBox {
	if len(data) < 4 {
		return nil
	}
	m := &MvhdBox{Version: data[0]}
	if m.Version == 0 {
		if len(data) < 20 {
			return nil
		}
		m.Timescale = binary.BigEndian.Uint32(data[12:16])
		m.Duration = uint64(binary.BigEndian.Uint32(data[16:20]))
	} else {
		if len(data) < 32 {
			return nil
		}
		m.Timescale = binary.BigEndian.Uint32(data[20:24])
		m.Duration = binary.BigEndian.Uint64(data[24:32])
	}
	return m
}

func parseTkhd(data []byte) *TkhdBox {
	if len(data) < 4 {
		return nil
	}
	t := &TkhdBox{Version: data[0]}
	if t.Version == 0 {
		if len(data) < 84 {
			return nil
		}
		t.TrackID = binary.BigEndian.Uint32(data[12:16])
		t.Duration = uint64(binary.BigEndian.Uint32(data[20:24]))
		t.Width = binary.BigEndian.Uint32(data[76:80])
		t.Height = binary.BigEndian.Uint32(data[80:84])
	} else {
		if len(data) < 96 {
			return nil
		}
		t.TrackID = binary.BigEndian.Uint32(data[20:24])
		t.Duration = binary.BigEndian.Uint64(data[28:36])
		t.Width = binary.BigEndian.Uint32(data[88:92])
		t.Height = binary.BigEndian.Uint32(data[92:96])
	}
	return t
}

func parseMdhd(data []byte) *MdhdBox {
	if len(data) < 4 {
		return nil
	}
	m := &MdhdBox{Version: data[0]}
	if m.Version == 0 {
		if len(data) < 20 {
			return nil
		}
		m.Timescale = binary.BigEndian.Uint32(data[12:16])
		m.Duration = uint64(binary.BigEndian.Uint32(data[16:20]))
	} else {
		if len(data) < 32 {
			return nil
		}
		m.Timescale = binary.BigEndian.Uint32(data[20:24])
		m.Duration = binary.BigEndian.Uint64(data[24:32])
	}
	return m
}

func parseHdlr(data []byte) *HdlrBox {
	if len(data) < 20 {
		return nil
	}
	h := &HdlrBox{
		HandlerType: string(data[8:12]),
	}
	if len(data) > 24 {
		name := data[24:]
		for i, b := range name {
			if b == 0 {
				name = name[:i]
				break
			}
		}
		h.Name = string(name)
	}
	return h
}

func parseStts(data []byte) *SttsBox {
	if len(data) < 8 {
		return nil
	}
	count := binary.BigEndian.Uint32(data[4:8])
	s := &SttsBox{Entries: make([]SttsEntry, 0, count)}
	for i := uint32(0); i < count; i++ {
		off := 8 + i*8
		if int(off+8) > len(data) {
			break
		}
		s.Entries = append(s.Entries, SttsEntry{
			SampleCount: binary.BigEndian.Uint32(data[off : off+4]),
			SampleDelta: binary.BigEndian.Uint32(data[off+4 : off+8]),
		})
	}
	return s
}

func parseStss(data []byte) *StssBox {
	if len(data) < 8 {
		return nil
	}
	count := binary.BigEndian.Uint32(data[4:8])
	s := &StssBox{SampleNumbers: make([]uint32, 0, count)}
	for i := uint32(0); i < count; i++ {
		off := 8 + i*4
		if int(off+4) > len(data) {
			break
		}
		s.SampleNumbers = append(s.SampleNumbers, binary.BigEndian.Uint32(data[off:off+4]))
	}
	return s
}

func parseStsz(data []byte) *StszBox {
	if len(data) < 12 {
		return nil
	}
	s := &StszBox{
		SampleSize:  binary.BigEndian.Uint32(data[4:8]),
		SampleCount: binary.BigEndian.Uint32(data[8:12]),
	}
	if s.SampleSize == 0 {
		s.EntrySizes = make([]uint32, 0, s.SampleCount)
		for i := uint32(0); i < s.SampleCount; i++ {
			off := 12 + i*4
			if int(off+4) > len(data) {
				break
			}
			s.EntrySizes = append(s.EntrySizes, binary.BigEndian.Uint32(data[off:off+4]))
		}
	}
	return s
}

func parseAvcC(data []byte) *AvcCBox {
	if len(data) < 6 {
		return nil
	}
	a := &AvcCBox{
		ProfileIndication:    data[1],
		ProfileCompatibility: data[2],
		LevelIndication:      data[3],
		LengthSizeMinusOne:   data[4] & 0x03,
	}
	numSPS := int(data[5] & 0x1f)
	off := 6
	for i := 0; i < numSPS; i++ {
		if off+2 > len(data) {
			break
		}
		spsLen := int(binary.BigEndian.Uint16(data[off : off+2]))
		off += 2
		if off+spsLen > len(data) {
			break
		}
		sps := make([]byte, spsLen)
		copy(sps, data[off:off+spsLen])
		a.SPS = append(a.SPS, sps)
		off += spsLen
	}
	if off >= len(data) {
		return a
	}
	numPPS := int(data[off])
	off++
	for i := 0; i < numPPS; i++ {
		if off+2 > len(data) {
			break
		}
		ppsLen := int(binary.BigEndian.Uint16(data[off : off+2]))
		off += 2
		if off+ppsLen > len(data) {
			break
		}
		pps := make([]byte, ppsLen)
		copy(pps, data[off:off+ppsLen])
		a.PPS = append(a.PPS, pps)
		off += ppsLen
	}
	return a
}

func parseAvc1(box *Box) *Avc1Box {
	if len(box.Data) < 28 {
		return nil
	}
	return &Avc1Box{
		Width:  binary.BigEndian.Uint16(box.Data[24:26]),
		Height: binary.BigEndian.Uint16(box.Data[26:28]),
	}
}

func parseMp4a(box *Box) *Mp4aBox {
	if len(box.Data) < 28 {
		return nil
	}
	return &Mp4aBox{
		ChannelCount: binary.BigEndian.Uint16(box.Data[16:18]),
		SampleSize:   binary.BigEndian.Uint16(box.Data[18:20]),
		SampleRate:   binary.BigEndian.Uint32(box.Data[24:28]) >> 16,
	}
}

// --- ESDS parsing ---

func parseEsds(data []byte) *EsdsBox {
	if len(data) < 5 {
		return nil
	}
	e := &EsdsBox{}
	// Skip version(1) + flags(3)
	parseESDescriptor(data[4:], e)
	return e
}

func parseESDescriptor(data []byte, e *EsdsBox) {
	if len(data) < 1 || data[0] != 0x03 {
		return
	}
	length, n := readDescLen(data[1:])
	pos := 1 + n
	end := pos + length
	if end > len(data) {
		end = len(data)
	}
	if pos+3 > end {
		return
	}
	flags := data[pos+2]
	pos += 3
	if flags&0x80 != 0 {
		pos += 2
	}
	if flags&0x40 != 0 && pos < end {
		urlLen := int(data[pos])
		pos += 1 + urlLen
	}
	if flags&0x20 != 0 {
		pos += 2
	}
	if pos < end {
		parseDecoderConfigDescriptor(data[pos:end], e)
	}
}

func parseDecoderConfigDescriptor(data []byte, e *EsdsBox) {
	if len(data) < 1 || data[0] != 0x04 {
		return
	}
	length, n := readDescLen(data[1:])
	pos := 1 + n
	end := pos + length
	if end > len(data) {
		end = len(data)
	}
	if pos+13 > end {
		return
	}
	e.ObjectTypeIndication = data[pos]
	e.StreamType = (data[pos+1] >> 2) & 0x3F
	e.MaxBitrate = binary.BigEndian.Uint32(data[pos+5 : pos+9])
	e.AvgBitrate = binary.BigEndian.Uint32(data[pos+9 : pos+13])
	pos += 13
	if pos < end {
		parseDecoderSpecificInfo(data[pos:end], e)
	}
}

func parseDecoderSpecificInfo(data []byte, e *EsdsBox) {
	if len(data) < 1 || data[0] != 0x05 {
		return
	}
	_, n := readDescLen(data[1:])
	pos := 1 + n
	if pos+2 > len(data) {
		return
	}
	e.AudioObjectType = (data[pos] >> 3) & 0x1F
	e.SamplingFreqIndex = ((data[pos] & 0x07) << 1) | ((data[pos+1] >> 7) & 0x01)
	e.ChannelConfig = (data[pos+1] >> 3) & 0x0F

	freqTable := [...]uint32{
		96000, 88200, 64000, 48000, 44100, 32000,
		24000, 22050, 16000, 12000, 11025, 8000, 7350,
	}
	if int(e.SamplingFreqIndex) < len(freqTable) {
		e.SamplingFrequency = freqTable[e.SamplingFreqIndex]
	}
}

func readDescLen(data []byte) (int, int) {
	length := 0
	n := 0
	for n < 4 && n < len(data) {
		b := data[n]
		n++
		length = (length << 7) | int(b&0x7f)
		if b&0x80 == 0 {
			break
		}
	}
	return length, n
}

// --- Name mappings ---

func h264ProfileName(indicator uint8) string {
	switch indicator {
	case 66:
		return "Baseline"
	case 77:
		return "Main"
	case 88:
		return "Extended"
	case 100:
		return "High"
	case 110:
		return "High 10"
	case 122:
		return "High 4:2:2"
	case 244:
		return "High 4:4:4 Predictive"
	default:
		return "Unknown"
	}
}

func aacProfileName(objectType uint8) string {
	switch objectType {
	case 1:
		return "Main"
	case 2:
		return "LC"
	case 3:
		return "SSR"
	case 4:
		return "LTP"
	case 5:
		return "SBR"
	case 29:
		return "PS"
	default:
		return "Unknown"
	}
}

func gcd(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
