package flv

import (
	"os"
	"fmt"
	"bytes"
)

type Header struct {
	Version      uint16
}

type Frame interface{
	ToBinary() ([]byte, error)
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

func (f CFrame) ToBinary() ([]byte, error) {
	d := make([]byte, 0)
	buf := bytes.NewBuffer(d)
	buf.WriteByte(byte(f.Type))
	bl := uint32(len(f.Body))
	buf.Write([]byte{byte(bl>>16), byte((bl>>8)&0xFF), byte(bl&0xFF)})
	buf.Write([]byte{byte((f.Dts>>16) & 0xFF), byte((f.Dts>>8) & 0xFF), byte(f.Dts & 0xFF)})
	buf.WriteByte(byte((f.Dts >> 24) & 0xFF))
	buf.Write(f.Body)
	buf.Write([]byte{byte((f.PrevTagSize>>24) & 0xFF), byte((f.PrevTagSize>>16) & 0xFF), byte((f.PrevTagSize>>8) & 0xFF), byte(f.PrevTagSize & 0xFF)})
	return buf.Bytes(), nil
}

type VideoFrame struct {
	CFrame
	CodecId     VideoCodec
	Width       uint16
	Height      uint16
}

func (f VideoFrame) ToBinary() ([]byte, error) {
	return f.ToBinary()
}

type AudioFrame struct {
	CFrame
	CodecId     AudioCodec
    Rate        AudioRate
    BitSize     AudioSize
	Channels    AudioType
}

func (f AudioFrame) ToBinary() ([]byte, error) {
	return f.ToBinary()
}

type MetaFrame struct {
	CFrame
}

func (f MetaFrame) ToBinary() ([]byte, error) {
	return f.ToBinary()
}

type FlvReader struct {
	inFile     *os.File
	width      uint16
	height     uint16
}

func NewReader(inFile *os.File) (*FlvReader) {
	return &FlvReader{
		inFile: inFile,
		width: 0,
		height: 0,
	}
}

func (frReader *FlvReader) ReadHeader() (*Header, error) {
	header := make([]byte, HEADER_LENGTH)
	_, err := frReader.inFile.Read(header)
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

	next_id := make([]byte, 4)
	_, err = frReader.inFile.Read(next_id)
	return &Header{Version: version}, nil
}

func (frReader *FlvReader) ReadFrame() (fr Frame, e error) {

	var n int
	var err error

	curPos, _ := frReader.inFile.Seek(0, os.SEEK_CUR)

	tagHeaderB := make([]byte, TAG_HEADER_LENGTH)
	n, err = frReader.inFile.Read(tagHeaderB)
	if TagSize(n) != TAG_HEADER_LENGTH {
		return nil, fmt.Errorf("bad record")
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
	n, err = frReader.inFile.Read(bodyBuf)
	if err != nil {
		return nil, err
	}

	prevTagSizeB := make([]byte, 4)
	n, err = frReader.inFile.Read(prevTagSizeB)
	if err != nil {
		return nil, err
	}
	prevTagSize := (uint32(prevTagSizeB[0]) << 24) | (uint32(prevTagSizeB[1]) << 16) | (uint32(prevTagSizeB[2]) << 8) | (uint32(prevTagSizeB[3]) << 0)


	pFrame := CFrame{
		Stream: stream,
		Dts: dts,
		Type: tagType,
		Position: curPos,
		Body: bodyBuf,
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
		rate := AudioRate((uint8(bodyBuf[0]) >> 2) & 0x03)
		bitSize := AudioSize((uint8(bodyBuf[0]) >> 1) & 0x01)
		channels := AudioType(uint8(bodyBuf[0]) & 0x01)
		resFrame = AudioFrame{CFrame: pFrame, CodecId: codecId, Rate: rate, BitSize: bitSize, Channels: channels}
	}


	return resFrame, nil
}
