package flv

import (
	"os"
	"fmt"
	"bytes"
)

type Header struct {
	Version      uint16
}

type Frame struct {
	Stream      uint32
	Dts         uint32
	Type        TagType
	Flavor      Flavor
	Position    int64
	Body        []byte
}

func ReadHeader(inFile *os.File) (*Header, error) {
	header := make([]byte, HEADER_LENGTH)
	_, err := inFile.Read(header)
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
	_, err = inFile.Read(next_id)
	return &Header{Version: version}, nil
}

func ReadTag(inFile *os.File) (fr *Frame, e error) {

	var n int
	var err error

	curPos, _ := inFile.Seek(0, os.SEEK_CUR)

	tagHeaderB := make([]byte, TAG_HEADER_LENGTH)
	n, err = inFile.Read(tagHeaderB)
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
	n, err = inFile.Read(bodyBuf)
	if err != nil {
		return nil, err
	}

	var flavor Flavor

	switch tagType {
	case TAG_TYPE_META:
		flavor = METADATA
	case TAG_TYPE_VIDEO:
		vft := VideoFrameType(uint(bodyBuf[0]) >> 4)
		switch vft {
		case VIDEO_FRAME_TYPE_KEYFRAME:
			flavor = KEYFRAME
		default:
			flavor = FRAME
		}
	case TAG_TYPE_AUDIO:
		flavor = FRAME
	}

	prevTagSizeB := make([]byte, 4)
	n, err = inFile.Read(prevTagSizeB)
	if err != nil {
		return nil, err
	}
	//prevTagSize := (uint32(prevTagSizeB[0]) << 24) | (uint32(prevTagSizeB[1]) << 16) | (uint32(prevTagSizeB[2]) << 8) | (uint32(prevTagSizeB[3]) << 0)

	return &Frame{
		Stream: stream,
		Dts: dts,
		Type: tagType,
		Flavor: flavor,
		Position: curPos,
		Body: bodyBuf,
	}, nil
}
