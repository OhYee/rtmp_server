package rtmp

// RTMP 类型常量
const (
	RTMPTypeSetChunkSize              = uint32(0x01)
	RTMPTypeACK                       = uint32(0x03)
	RTMPTypeUserControlMessage        = uint32(0x04)
	RTMPTypeWindowAcknowledgementSize = uint32(0x05)
	RTMPTypeSetPeerBandwidth          = uint32(0x06)
	RTMPTypeAudioData                 = uint32(0x08)
	RTMPTypeVideoData                 = uint32(0x09)
	RTMPTypeAMF3Command               = uint32(0x11)
	RTMPTypeAMFData                   = uint32(0x12)
	RTMPTypeAMF0Command               = uint32(0x14)
	RTMPTypeStreamData                = uint32(0x16)
)

// RTMP AMF Command 常量字段
const (
	AMFCommandName                  = "Command Name"
	AMFCommandTransactionID         = "Transaction ID"
	AMFCommandObject                = "Command Object"
	AMFCommandOptionalUserArguments = "Optional User Arguments"
)

// RTMP 其他 常量字段
const (
	AMFType  = "Type"
	AMFValue = "Value"
)

// 用户控制信息 常量字段
const (
	UserControlMessageSetBufferLength = uint32(3)
)
