package rtmp

import (
	"encoding/binary"
	"log"

	c "../lib/colorful"
	s "../server"
	"github.com/pkg/errors"
)

/*

一个 RTMP 连接

*/

// Connect 连接对象
type Connect struct {
	WithinServer *Server    // 所在的RTMP服务
	WithinStream *Stream    // 所在的流信息
	Conn         *s.Connect // 服务连接
	Close        bool       // 是否需要关闭

	RecvChunkSize                 uint32 // 对方的最大Chunk长度
	SendChunkSize                 uint32 // 本地的最大Chunk长度
	RecvWindowAcknowledgementSize uint32 // 对方的窗口长度
	SendWindowAcknowledgementSize uint32 // 本地的窗口长度
	RecvBandwidth                 uint32 // 带宽大小
	RecvBandwidthType             uint32 // 带宽类型
	SendBandwidth                 uint32 // 带宽大小
	BufferSize                    uint32 // 缓冲区大小

	TotalReceive uint32 // 已接受的总数
	Seq          uint32 // 窗口计数

	LastRecvChunk  map[uint32]Chunk
	LastSendChunk  map[uint32]Chunk
	RecvMessageMap map[uint32]*Message // 未完全接收的Message列表

	isBegin   bool   // Tag是否已发送
	beginTime uint32 // 开始关键帧时间戳
	TotalTime uint32 // 视频流时间戳总时间

	FullName   string
	AppName    string
	StreamName string

	VideoChunkID uint32
	AudioChunkID uint32
	StreamID     uint32

	Thread AVThread

	Test bool
}

// NewConnect Connect构造函数
func NewConnect(conn *s.Connect, server *Server) *Connect {
	connect := Connect{}
	connect.Conn = conn
	connect.WithinServer = server

	connect.LastRecvChunk = make(map[uint32]Chunk)
	connect.LastSendChunk = make(map[uint32]Chunk)

	connect.RecvChunkSize = 128
	connect.SendChunkSize = 128
	connect.RecvMessageMap = make(map[uint32]*Message)

	connect.VideoChunkID = 60
	connect.AudioChunkID = 61
	connect.StreamID = 67

	connect.isBegin = false
	connect.beginTime = 0
	connect.Thread = NewAVThread(&connect)

	connect.Test = false

	return &connect
}

// Server RTMP连接服务
func (conn *Connect) Server() error {
	// 握手
	err := Handshake(conn)
	if err != nil {
		return errors.WithStack(err)
	}
	log.Println(c.Front("Handshake ok.", c.G))

	if err := conn.loop(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// loop 循环过程
func (conn *Connect) loop() error {
	for {
		msg, err := NewMessage(conn)
		if err != nil {
			return errors.WithStack(err)
		}

		if err := msg.Solve(conn); err != nil {
			return errors.WithStack(err)
		}

		if conn.Close {
			// 连接结束
			break
		}
	}
	return nil
}

// CloseServer 关闭连接
func (conn *Connect) CloseServer() {
	conn.Close = true
}

// BeforeClose 关闭连接前的处理函数，与CloseServer不同，这里包括报错关闭的情况
func (conn *Connect) BeforeClose() {
	if conn.WithinStream != nil {
		conn.WithinStream.DelConnect(conn)
	}
	conn.Thread.Stop()
}

// Read 读入指定长度的数据
func (conn *Connect) Read(len uint32) ([]byte, error) {
	var readLength uint32
	data := make([]byte, len)

	for readLength < len {
		buff := make([]byte, len-readLength)
		l, err := conn.Conn.Read(buff)
		if err != nil {
			// log.Println(err)
			return data, err
			// continue
		}
		if conn.Close {
			return data, errors.New("Close Server")
		}
		data = append(data[:readLength], buff[:l]...)
		readLength += uint32(l)
	}

	return data, nil
}

// Write 写出数据
func (conn *Connect) Write(b []byte) (int, error) {
	return conn.Conn.Write(b)
}

// WriteChunk 写出Chunk
func (conn *Connect) WriteChunk(chk Chunk) (int, error) {
	conn.LastSendChunk[chk.Basic.ChunkStreamID] = chk
	b, err := chk.Bytes()
	if err != nil {
		return 0, errors.WithStack(err)
	}
	// log.Println(c.Front("%v Write chunk %v", c.Y, conn, b))
	return conn.Conn.Write(b)
}

// WriteChunks 写出Chunks
func (conn *Connect) WriteChunks(chunks []Chunk) error {
	for _, chk := range chunks {
		_, err := conn.WriteChunk(chk)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// WriteMessage 写出Message
func (conn *Connect) WriteMessage(msg Message) error {
	chunks := msg.ToChunks(conn)
	err := conn.WriteChunks(chunks)
	if err != nil {
		return errors.WithStack(err)
	}
	// log.Println(c.Front("%v %d", c.Y, msg.Type, msg.Length))
	return err
}

// SendACK 发送一个窗口确认消息
func (conn *Connect) SendACK() error {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, uint32(conn.TotalReceive))
	msg, err := MakeMessage(RTMPTypeACK, data, 0, 0, 0)
	if err != nil {
		err = conn.WriteMessage(msg)
	}
	if err != nil {
		return errors.WithStack(err)
	}
	return err
}

// SendResponse 发送命令的响应消息
func (conn *Connect) SendResponse(amfCommand AMFCommand, streamID uint32, chunkStreamID uint32) error {
	msg, err := MakeMessage(
		RTMPTypeAMF0Command,
		amfCommand,
		streamID,
		chunkStreamID,
		0,
	)
	if err != nil {
		return errors.WithStack(err)
	}
	err = conn.WriteMessage(msg)
	if err != nil {
		return errors.WithStack(err)
	}
	if err != nil {
		return errors.WithStack(err)
	}
	return err
}

// SendStreamIsRecord 发送流记录命令
func (conn *Connect) SendStreamIsRecord(streamID uint32) error {
	_, err := conn.Write([]byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, byte(streamID)})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// SendStreamBegin 发送流开始命令
func (conn *Connect) SendStreamBegin(streamID uint32) error {
	_, err := conn.Write([]byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, byte(streamID)})
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// SendWinACKSize 发送窗口大小命令
func (conn *Connect) SendWinACKSize(size uint32) error {
	conn.SendWindowAcknowledgementSize = size
	sizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBytes, conn.SendWindowAcknowledgementSize)

	msg, err := MakeMessage(
		RTMPTypeWindowAcknowledgementSize,
		sizeBytes,
		0,
		2,
		0,
	)
	err = conn.WriteMessage(msg)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// SendSetPeerBandwidth 发送窗口大小命令
func (conn *Connect) SendSetPeerBandwidth(size uint32) error {
	conn.SendBandwidth = size
	sizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBytes, conn.SendBandwidth)

	msg, err := MakeMessage(
		RTMPTypeSetPeerBandwidth,
		append(sizeBytes, 0x02),
		0,
		2,
		0,
	)
	err = conn.WriteMessage(msg)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// SendSetChunkSize 设置分块大小
func (conn *Connect) SendSetChunkSize(size uint32) error {
	conn.SendChunkSize = size
	sizeBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sizeBytes, conn.SendChunkSize)

	msg, err := MakeMessage(
		RTMPTypeSetChunkSize,
		sizeBytes,
		0,
		2,
		0,
	)
	err = conn.WriteMessage(msg)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// SendAVMessage 发送音视频消息
func (conn *Connect) SendAVMessage(msg Message) error {

	if !conn.isBegin {
		// 尚未发送关键帧

		frameType := uint32(msg.Data[0]) >> 4
		if msg.Type == RTMPTypeAudioData || (frameType != 1 && frameType != 4) {
			// 等待有视频帧发送过后再发送音频帧
			// 非关键帧也应该忽略
			log.Println("F")
			return nil
		}
		tag, ok := conn.WithinStream.GetTag()

		if !ok {
			// 推流端还未发送关键帧
			return nil
		}
		tag.Timestamp = 0
		tag.ChunkStreamID = conn.VideoChunkID
		conn.isBegin = true
		conn.beginTime = msg.Timestamp

		err := conn.WriteMessage(tag)

		if err != nil {
			return errors.WithStack(err)
		}
	}

	var csid uint32
	if msg.Type == RTMPTypeVideoData {
		csid = conn.VideoChunkID
	} else {
		csid = conn.AudioChunkID
	}
	frame := msg.Copy()
	frame.ChunkStreamID = csid

	frame.Timestamp -= conn.beginTime

	err := conn.WriteMessage(frame)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
