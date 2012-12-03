package flv

import (
	"bytes"
	"fmt"
	"github.com/metachord/amf.go/amf0"
	"io"
	"os"
)

type Header struct {
	Version uint16
	Body    []byte
}

type Frame interface {
	WriteFrame(io.Writer) error
	GetStream() uint32
	GetDts() uint32
	String() string
}

type CFrame struct {
	Stream      uint32
	Dts         uint32
	Type        TagType
	Flavor      Flavor
	Position    int64
	Body        []byte
	PrevTagSize uint32
}

type VideoFrame struct {
	CFrame
	CodecId VideoCodec
	Width   uint16
	Height  uint16
}

type AudioFrame struct {
	CFrame
	CodecId  AudioCodec
	Rate     uint32
	BitSize  AudioSize
	Channels AudioType
}

type MetaFrame struct {
	CFrame
}

func (f CFrame) WriteFrame(w io.Writer) error {
	bl := uint32(len(f.Body))
	var err error
	err = writeType(w, f.Type)
	if err != nil {
		return err
	}
	err = writeBodyLength(w, bl)
	err = writeDts(w, f.Dts)
	err = writeStream(w, f.Stream)
	err = writeBody(w, f.Body)
	prevTagSize := bl + uint32(TAG_HEADER_LENGTH)
	err = writePrevTagSize(w, prevTagSize)
	return nil
}

func (f CFrame) GetStream() uint32 {
	return f.Stream
}
func (f CFrame) GetDts() uint32 {
	return f.Dts
}

func writeType(w io.Writer, t TagType) error {
	_, err := w.Write([]byte{byte(t)})
	return err
}

func writeBodyLength(w io.Writer, bl uint32) error {
	_, err := w.Write([]byte{byte(bl >> 16), byte((bl >> 8) & 0xFF), byte(bl & 0xFF)})
	return err
}

func writeDts(w io.Writer, dts uint32) error {
	_, err := w.Write([]byte{byte((dts >> 16) & 0xFF), byte((dts >> 8) & 0xFF), byte(dts & 0xFF)})
	if err != nil {
		return err
	}
	_, err = w.Write([]byte{byte((dts >> 24) & 0xFF)})
	return err
}

func writeStream(w io.Writer, stream uint32) error {
	_, err := w.Write([]byte{byte(stream >> 16), byte((stream >> 8) & 0xFF), byte(stream & 0xFF)})
	return err
}

func writeBody(w io.Writer, body []byte) error {
	_, err := w.Write(body)
	return err
}

func writePrevTagSize(w io.Writer, prevTagSize uint32) error {
	_, err := w.Write([]byte{byte((prevTagSize >> 24) & 0xFF), byte((prevTagSize >> 16) & 0xFF), byte((prevTagSize >> 8) & 0xFF), byte(prevTagSize & 0xFF)})
	return err
}

func (f VideoFrame) String() string {
	return fmt.Sprintf("%d\t%d\t%s\t%s\t%d bytes\t%dx%d", f.CFrame.Stream, f.CFrame.Dts, f.CFrame.Type, f.CodecId, len(f.CFrame.Body), f.Width, f.Height)
}

func (f AudioFrame) String() string {
	return fmt.Sprintf("%d\t%d\t%s\t%s\t{%d,%s,%s}", f.CFrame.Stream, f.CFrame.Dts, f.CFrame.Type, f.CodecId, f.Rate, f.BitSize, f.Channels)
}

func (f MetaFrame) String() string {
	buf := bytes.NewReader(f.CFrame.Body)
	dec := amf0.NewDecoder(buf)
	evName, err := dec.Decode()
	mds := ""
	if err == nil {
		switch evName {
		case amf0.StringType("onMetaData"):
			md, err := dec.Decode()
			if err == nil {

				var ea map[amf0.StringType]interface{}
				switch md := md.(type) {
				case *amf0.EcmaArrayType:
					ea = *md
				case *amf0.ObjectType:
					ea = *md
				}
				for k, v := range ea {
					mds += fmt.Sprintf("%v=%+v;", k, v)
				}
			}
		}
	}

	return fmt.Sprintf("%d\t%d\t%s\t%s", f.CFrame.Stream, f.CFrame.Dts, f.CFrame.Type, mds)
}

type FlvReader struct {
	InFile *os.File
	width  uint16
	height uint16
}

func NewReader(inFile *os.File) *FlvReader {
	return &FlvReader{
		InFile: inFile,
		width:  0,
		height: 0,
	}
}

type FlvWriter struct {
	OutFile *os.File
}

func NewWriter(outFile *os.File) *FlvWriter {
	return &FlvWriter{
		OutFile: outFile,
	}
}

func (frReader *FlvReader) ReadHeader() (*Header, error) {
	header := make([]byte, HEADER_LENGTH+4)
	_, err := frReader.InFile.Read(header)
	if err != nil {
		return nil, err
	}

	sig := header[0:3]
	if bytes.Compare(sig, []byte(SIG)) != 0 {
		return nil, fmt.Errorf("bad file format")
	}
	version := (uint16(header[3]) << 8) | (uint16(header[4]) << 0)
	//skip := header[4:5]
	//offset := header[5:9]
	//next_id := header[9:13]

	return &Header{Version: version, Body: header}, nil
}

func (frWriter *FlvWriter) WriteHeader(header *Header) error {
	_, err := frWriter.OutFile.Write(header.Body)
	if err != nil {
		return err
	}
	return nil
}

func (frReader *FlvReader) ReadFrame() (fr Frame, e error) {

	var n int
	var err error

	curPos, _ := frReader.InFile.Seek(0, os.SEEK_CUR)

	tagHeaderB := make([]byte, TAG_HEADER_LENGTH)
	n, err = frReader.InFile.Read(tagHeaderB)
	if n == 0 {
		return nil, nil
	}
	if TagSize(n) != TAG_HEADER_LENGTH {
		return nil, fmt.Errorf("bad tag length: %d", n)
	}
	if err != nil {
		return nil, err
	}

	tagType := TagType(tagHeaderB[0])
	bodyLen := (uint32(tagHeaderB[1]) << 16) | (uint32(tagHeaderB[2]) << 8) | (uint32(tagHeaderB[3]) << 0)
	ts := (uint32(tagHeaderB[4]) << 16) | (uint32(tagHeaderB[5]) << 8) | (uint32(tagHeaderB[6]) << 0)
	tsExt := uint32(tagHeaderB[7])
	stream := (uint32(tagHeaderB[8]) << 16) | (uint32(tagHeaderB[9]) << 8) | (uint32(tagHeaderB[10]) << 0)

	var dts uint32
	dts = (tsExt << 24) | ts

	bodyBuf := make([]byte, bodyLen)
	n, err = frReader.InFile.Read(bodyBuf)
	if err != nil {
		return nil, err
	}

	prevTagSizeB := make([]byte, PREV_TAG_SIZE_LENGTH)
	n, err = frReader.InFile.Read(prevTagSizeB)
	if err != nil {
		return nil, err
	}
	prevTagSize := (uint32(prevTagSizeB[0]) << 24) | (uint32(prevTagSizeB[1]) << 16) | (uint32(prevTagSizeB[2]) << 8) | (uint32(prevTagSizeB[3]) << 0)

	pFrame := CFrame{
		Stream:      stream,
		Dts:         dts,
		Type:        tagType,
		Position:    curPos,
		Body:        bodyBuf,
		PrevTagSize: prevTagSize,
	}

	var resFrame Frame

	switch tagType {
	case TAG_TYPE_META:
		pFrame.Flavor = METADATA
		resFrame = MetaFrame{pFrame}
	case TAG_TYPE_VIDEO:
		vft := VideoFrameType(uint8(bodyBuf[0]) >> 4)
		codecId := VideoCodec(uint8(bodyBuf[0]) & 0x0F)
		switch vft {
		case VIDEO_FRAME_TYPE_KEYFRAME:
			pFrame.Flavor = KEYFRAME
			switch codecId {
			case VIDEO_CODEC_ON2VP6:
				hHelper := (uint16(bodyBuf[1]) >> 4) & 0x0F
				wHelper := uint16(bodyBuf[1]) & 0x0F
				w := uint16(bodyBuf[5])
				h := uint16(bodyBuf[6])

				frReader.width = w*16 - wHelper
				frReader.height = h*16 - hHelper
			}
		default:
			pFrame.Flavor = FRAME
		}
		resFrame = VideoFrame{CFrame: pFrame, CodecId: codecId, Width: frReader.width, Height: frReader.height}
	case TAG_TYPE_AUDIO:
		pFrame.Flavor = FRAME
		codecId := AudioCodec(uint8(bodyBuf[0]) >> 4)
		rate := audioRate(AudioRate((uint8(bodyBuf[0]) >> 2) & 0x03))
		bitSize := AudioSize((uint8(bodyBuf[0]) >> 1) & 0x01)
		channels := AudioType(uint8(bodyBuf[0]) & 0x01)
		resFrame = AudioFrame{CFrame: pFrame, CodecId: codecId, Rate: rate, BitSize: bitSize, Channels: channels}
	}

	return resFrame, nil
}

func audioRate(ar AudioRate) uint32 {
	var ret uint32
	switch ar {
	case AUDIO_RATE_5_5:
		ret = 5500
	case AUDIO_RATE_11:
		ret = 11000
	case AUDIO_RATE_22:
		ret = 22000
	case AUDIO_RATE_44:
		ret = 44000
	}
	return ret
}

func (frWriter *FlvWriter) WriteFrame(fr Frame) (e error) {
	return fr.WriteFrame(frWriter.OutFile)
}
