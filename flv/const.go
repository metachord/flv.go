package flv


type TagSize uint
const (
	HEADER_LENGTH         TagSize =  9
	PREV_TAG_SIZE_LENGTH  TagSize =  4
	TAG_HEADER_LENGTH     TagSize =  11
)

type TagType uint
const (
	TAG_TYPE_AUDIO TagType       =  8
	TAG_TYPE_VIDEO TagType       =  9
	TAG_TYPE_META  TagType       =  18
)

type VideoFrameType uint
const (
	VIDEO_FRAME_TYPE_KEYFRAME         VideoFrameType  = 1
	VIDEO_FRAME_TYPEINTER_FRAME       VideoFrameType  = 2
	VIDEO_FRAME_TYPEDISP_INTER_FRAME  VideoFrameType  = 3
	VIDEO_FRAME_TYPE_GENERATED        VideoFrameType  = 4
	VIDEO_FRAME_TYPE_COMMAND          VideoFrameType  = 5
)

type Flavor uint
const (
	METADATA      Flavor = iota
	FRAME
	KEYFRAME
)
