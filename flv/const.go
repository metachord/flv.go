package flv

const (
	SIG   = "FLV"
)

type TagSize byte
const (
	HEADER_LENGTH         TagSize = 9
	PREV_TAG_SIZE_LENGTH  TagSize = 4
	TAG_HEADER_LENGTH     TagSize = 11
)

type TagType byte
const (
	TAG_TYPE_AUDIO    TagType = 8
	TAG_TYPE_VIDEO    TagType = 9
	TAG_TYPE_META     TagType = 18
)

type VideoFrameType byte
const (
	VIDEO_FRAME_TYPE_KEYFRAME         VideoFrameType = 1
	VIDEO_FRAME_TYPEINTER_FRAME       VideoFrameType = 2
	VIDEO_FRAME_TYPEDISP_INTER_FRAME  VideoFrameType = 3
	VIDEO_FRAME_TYPE_GENERATED        VideoFrameType = 4
	VIDEO_FRAME_TYPE_COMMAND          VideoFrameType = 5
)

type VideoCodec byte
const (
	VIDEO_CODEC_JPEG           VideoCodec = 1
	VIDEO_CODEC_SORENSON       VideoCodec = 2
	VIDEO_CODEC_SCREENVIDEO    VideoCodec = 3
	VIDEO_CODEC_ON2VP6         VideoCodec = 4
	VIDEO_CODEC_ON2VP6_ALPHA   VideoCodec = 5
	VIDEO_CODEC_SCREENVIDEO2   VideoCodec = 6
	VIDEO_CODEC_AVC            VideoCodec = 7
)

type VideoAvc byte
const (
	VIDEO_AVC_SEQUENCE_HEADER  VideoAvc = 0
	VIDEO_AVC_NALU             VideoAvc = 1
	VIDEO_AVC_SEQUENCE_END     VideoAvc = 2
)

type AudioType byte
const (
	AUDIO_TYPE_MONO     AudioType = 0
	AUDIO_TYPE_STEREO   AudioType = 1
)

type AudioSize byte
const (
	AUDIO_SIZE_8BIT   AudioSize = 0
	AUDIO_SIZE_16BIT  AudioSize = 1
)


type AudioRate byte
const (
	AUDIO_RATE_5_5  AudioRate = 0
	AUDIO_RATE_11   AudioRate = 1
	AUDIO_RATE_22   AudioRate = 2
	AUDIO_RATE_44   AudioRate = 3
)


type AudioCodec byte
const (
	AUDIO_CODEC_PCM          AudioCodec = 0
	AUDIO_CODEC_ADPCM        AudioCodec = 1
	AUDIO_CODEC_MP3          AudioCodec = 2
	AUDIO_CODEC_PCM_LE       AudioCodec = 3
	AUDIO_CODEC_NELLYMOSER8  AudioCodec = 5
	AUDIO_CODEC_NELLYMOSER   AudioCodec = 6
	AUDIO_CODEC_A_G711       AudioCodec = 7
	AUDIO_CODEC_MU_G711      AudioCodec = 8
	AUDIO_CODEC_RESERVED     AudioCodec = 9
	AUDIO_CODEC_AAC          AudioCodec = 10
	AUDIO_CODEC_SPEEX        AudioCodec = 11
	AUDIO_CODEC_MP3_8KHZ     AudioCodec = 14
	AUDIO_CODEC_DEVICE       AudioCodec = 15
)

type AudioAac byte
const (
	AUDIO_AAC_SEQUENCE_HEADER  AudioAac = 0
	AUDIO_AAC_RAW              AudioAac = 1
)

type Flavor byte
const (
	METADATA      Flavor = iota
	FRAME
	KEYFRAME
)
