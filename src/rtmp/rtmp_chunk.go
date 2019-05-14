package rtmp

import (
	"bytes"
	"encoding/binary"

	"../lib"
	"github.com/pkg/errors"
)

// Chunk ...
type Chunk struct {
	Basic   BasicHeader
	Message MessageHeader
	Data    []byte
}

// NewChunk 读入一个新的chunk
func NewChunk(conn *Connect) (Chunk, error) {
	chk := Chunk{}

	// 读入头部
	err := chk.Basic.Read(conn)
	if err != nil {
		return chk, errors.WithStack(err)
	}
	err = chk.Message.Read(chk.Basic.Format, chk.Basic.ChunkStreamID, conn)
	if err != nil {
		return chk, errors.WithStack(err)
	}

	// 判断要读入的长度
	msid := chk.Message.MessageStreamID
	mslen := chk.Message.MessageLength
	msg, ok := conn.RecvMessageMap[msid]
	var readLength uint32
	if ok {
		readLength = lib.Min(msg.Length-msg.ReadLength, conn.RecvChunkSize)
	} else {
		readLength = lib.Min(mslen, conn.RecvChunkSize)
	}

	// 读入数据
	chk.Data, err = conn.Read(readLength)
	if err != nil {
		return chk, errors.WithStack(err)
	}

	// log.Printf(c.Front("Chunk Received %+v", c.S, chk))
	return chk, nil
}

// Bytes 输出chunk的数据
func (chk *Chunk) Bytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	numByte := make([]byte, 8)

	switch chk.Basic.Format {
	case 0:
		// 写出BasicHeader
		binary.BigEndian.PutUint16(numByte, uint16(0x3f&chk.Basic.ChunkStreamID))
		buf.Write(numByte[1:2])
		// 写出时间戳
		binary.BigEndian.PutUint32(numByte, uint32(chk.Message.Timestamp))
		buf.Write(numByte[1:4])
		// 写出长度
		binary.BigEndian.PutUint32(numByte, uint32(chk.Message.MessageLength))
		buf.Write(numByte[1:4])
		// 写出类型
		binary.BigEndian.PutUint16(numByte, uint16(chk.Message.MessageType))
		buf.Write(numByte[1:2])
		// 写出流id
		binary.LittleEndian.PutUint32(numByte, uint32(chk.Message.MessageStreamID))
		buf.Write(numByte[0:4])
	case 1:
		// 写出BasicHeader
		binary.BigEndian.PutUint16(numByte, uint16(0x40|(0x3f&chk.Basic.ChunkStreamID)))
		buf.Write(numByte[1:2])
		// 写出时间戳
		binary.BigEndian.PutUint32(numByte, uint32(chk.Message.Timestamp))
		buf.Write(numByte[1:4])
		// 写出长度
		binary.BigEndian.PutUint32(numByte, uint32(chk.Message.MessageLength))
		buf.Write(numByte[1:4])
		// 写出类型
		binary.BigEndian.PutUint16(numByte, uint16(chk.Message.MessageType))
		buf.Write(numByte[1:2])
	case 2:
		// 写出BasicHeader
		binary.BigEndian.PutUint16(numByte, uint16(0x80|(0x3f&chk.Basic.ChunkStreamID)))
		buf.Write(numByte[1:2])
		// 写出时间戳
		binary.BigEndian.PutUint32(numByte, uint32(chk.Message.Timestamp))
		buf.Write(numByte[1:4])
		// 写出长度
		binary.BigEndian.PutUint32(numByte, uint32(chk.Message.MessageLength))
		buf.Write(numByte[1:4])
	case 3:
		// 写出BasicHeader
		binary.BigEndian.PutUint16(numByte, uint16(0xC0|(0x3f&chk.Basic.ChunkStreamID)))
		buf.Write(numByte[1:2])
	default:
		return make([]byte, 0), errors.WithStack(errors.New("RTMP chunk basic header format type error"))
	}
	return lib.ByteArrayConcat(buf.Bytes(), chk.Data), nil
}

/*

Basic Header

*/

// BasicHeader RTMP基础头部
type BasicHeader struct {
	Format        uint32
	ChunkStreamID uint32
}

// Read 读取Basic Header
func (basic *BasicHeader) Read(conn *Connect) error {
	data, err := conn.Read(1)
	if err != nil {
		return errors.WithStack(err)
	}

	csid := uint32(data[0]) & 0x3f
	basic.Format = uint32((int(data[0]) & 0xc0) >> 6)
	if csid == 0 {
		// 格式1
		data, err = conn.Read(1)
		if err != nil {
			return errors.WithStack(err)
		}
		basic.ChunkStreamID = uint32(int(data[0]) + 64)
	} else if csid == 1 {
		// 格式2
		data, err = conn.Read(2)
		if err != nil {
			return errors.WithStack(err)
		}
		basic.ChunkStreamID = uint32(int(data[1])*256 + int(data[0]) + 64)
	} else {
		// 格式0
		basic.ChunkStreamID = csid
	}
	return nil
}

/*

MessageHeader

*/

// MessageHeader ...
type MessageHeader struct {
	Timestamp       uint32
	MessageLength   uint32
	MessageType     uint32
	MessageStreamID uint32
}

// Read 读取Message Header
func (basic *MessageHeader) Read(format uint32, csid uint32, conn *Connect) error {
	data := make([]byte, 16)
	var err error

	lastReceChunk, ok := conn.LastRecvChunk[csid]
	if !ok {
		lastReceChunk = Chunk{}
	}

	switch format {
	case 0:
		data, err = conn.Read(11)
		if err != nil {
			return errors.WithStack(err)
		}
		basic.Timestamp = lib.ToUint32(data[0:3])
		basic.MessageLength = lib.ToUint32(data[3:6])
		basic.MessageType = lib.ToUint32(data[6:7])
		basic.MessageStreamID = lib.ToUint32(data[7:])
		if basic.Timestamp == 0xffffff {
			data, err = conn.Read(4)
			if err != nil {
				return errors.WithStack(err)
			}
			basic.Timestamp = lib.ToUint32(data[0:4])
		}
	case 1:
		data, err = conn.Read(7)
		if err != nil {
			return errors.WithStack(err)
		}
		basic.Timestamp = lib.ToUint32(data[0:3]) + lastReceChunk.Message.Timestamp
		basic.MessageLength = lib.ToUint32(data[3:6])
		basic.MessageType = lib.ToUint32(data[6:7])
		basic.MessageStreamID = lastReceChunk.Message.MessageStreamID
	case 2:
		data, err = conn.Read(3)
		if err != nil {
			return errors.WithStack(err)
		}
		basic.Timestamp = lib.ToUint32(data[0:3]) + lastReceChunk.Message.Timestamp
		basic.MessageLength = lastReceChunk.Message.MessageLength
		basic.MessageType = lastReceChunk.Message.MessageType
		basic.MessageStreamID = lastReceChunk.Message.MessageStreamID
	case 3:
		data, err = conn.Read(0)
		if err != nil {
			return errors.WithStack(err)
		}
		basic.Timestamp = lastReceChunk.Message.Timestamp
		basic.MessageLength = lastReceChunk.Message.MessageLength
		basic.MessageType = lastReceChunk.Message.MessageType
		basic.MessageStreamID = lastReceChunk.Message.MessageStreamID
	default:
		return errors.WithStack(errors.New("RTMP format error"))
	}
	return nil
}
