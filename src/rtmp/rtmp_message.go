package rtmp

import (
	"bytes"
	"fmt"
	"log"

	"../lib"
	c "../lib/colorful"
	"./amf"
	"github.com/pkg/errors"
)

// Message RTMP Message
type Message struct {
	Timestamp     uint32
	Type          uint32
	Length        uint32
	ReadLength    uint32
	StreamID      uint32
	ChunkStreamID uint32
	Data          []byte
}

// NewMessage 读入一条message
func NewMessage(conn *Connect) (Message, error) {
	chunk, err := NewChunk(conn)
	if err != nil {
		return Message{}, errors.WithStack(err)
	}
	conn.Seq++
	if conn.Seq >= conn.RecvWindowAcknowledgementSize {
		conn.TotalReceive += conn.Seq
		conn.Seq = 0
		conn.SendACK()
	}

	msid := chunk.Message.MessageStreamID

	msg, ok := conn.RecvMessageMap[msid]
	if ok {
		// Map中存在,更新该msg
		msg.Data = lib.ByteArrayConcat(msg.Data, chunk.Data)
		msg.ReadLength += uint32(len(chunk.Data))
		if msg.ReadLength >= msg.Length {
			// 读取完毕,从Map删除
			delete(conn.RecvMessageMap, msid)
		}
	} else {
		// Map中不存在，生成一个新的Message
		msg = &Message{
			Length:        chunk.Message.MessageLength,
			ReadLength:    uint32(len(chunk.Data)),
			Data:          chunk.Data,
			Type:          chunk.Message.MessageType,
			ChunkStreamID: chunk.Basic.ChunkStreamID,
			StreamID:      chunk.Message.MessageStreamID,
			Timestamp:     chunk.Message.Timestamp,
		}

		if msg.ReadLength < msg.Length {
			// 未读取完毕,插入到Map中
			conn.RecvMessageMap[msid] = msg
		}
	}
	conn.LastRecvChunk[chunk.Basic.ChunkStreamID] = chunk

	if msg.ReadLength >= msg.Length {
		return *msg, nil
	}
	return NewMessage(conn)
}

/*
	Message 数据写出部分
*/

// MakeMessage 封装一个包，返回生成的Chunk列表
func MakeMessage(messageType uint32, data interface{}, streamID uint32, chunkStreamID uint32, ts uint32) (Message, error) {
	msg := Message{}
	msg.Timestamp = ts
	msg.Type = messageType
	msg.StreamID = streamID
	msg.ChunkStreamID = chunkStreamID
	var err error

	// chk := Chunk{}
	switch messageType {
	case RTMPTypeAMF0Command:
		fallthrough
	case RTMPTypeAMF3Command:
		dataCommand, ok := data.(AMFCommand)
		if ok != true {
			return msg, fmt.Errorf("RTMP 类型错误 %+v(expect string, actual %v)", data, data)
		}
		msg.Data, err = dataCommand.ToBytes()
		if err != nil {
			return msg, errors.WithStack(err)
		}
	default:
		byteData, ok := data.([]byte)
		if ok {
			msg.Data = byteData
		}
	}
	msg.Length = uint32(len(msg.Data))
	msg.ReadLength = msg.Length

	return msg, nil
}

// ToChunks 将Message转换为Chunk用于发送
func (msg *Message) ToChunks(conn *Connect) []Chunk {
	chunkList := make([]Chunk, 0)

	var i uint32
	for i < msg.Length {
		chk := Chunk{}
		if i == 0 {
			chk.Basic.Format = 0
			chk.Basic.ChunkStreamID = msg.ChunkStreamID
			chk.Message.MessageLength = msg.Length
			chk.Message.Timestamp = msg.Timestamp
			chk.Message.MessageStreamID = msg.StreamID
			chk.Message.MessageType = msg.Type
		} else {
			chk.Basic.Format = 3
			chk.Basic.ChunkStreamID = msg.ChunkStreamID
		}
		chk.Data = msg.Data[i:lib.Min(i+conn.SendChunkSize, msg.Length)]
		chunkList = append(chunkList, chk)

		i = i + conn.SendChunkSize
	}
	return chunkList
}

// Copy 拷贝Message的副本
func (msg *Message) Copy() Message {
	return Message{
		Timestamp:  msg.Timestamp,
		Type:       msg.Type,
		Length:     msg.Length,
		ReadLength: msg.ReadLength,
		StreamID:   msg.StreamID,
		Data:       msg.Data,
	}
}

// AMFCommand AMF命令格式
type AMFCommand struct {
	CommandName           string
	TransactionID         float64
	CommandObject         interface{}
	OptionalUserArguments interface{}
}

// ToBytes 将AMFCommand转换为字节流
func (amfCommand *AMFCommand) ToBytes() ([]byte, error) {
	buf := new(bytes.Buffer)

	buf.Write(amf.NewString(amfCommand.CommandName).Bytes())
	buf.Write(amf.NewNumber(amfCommand.TransactionID).Bytes())

	amfData, err := amf.MakeAMF(amfCommand.CommandObject)
	if err != nil {
		return buf.Bytes(), errors.WithStack(err)
	}
	buf.Write(amfData.Bytes())

	amfData, err = amf.MakeAMF(amfCommand.OptionalUserArguments)
	if err != nil {
		return buf.Bytes(), errors.WithStack(err)
	}
	buf.Write(amfData.Bytes())

	return buf.Bytes(), nil
}

/*
	消息处理
*/

// Solve 消息处理函数
func (msg *Message) Solve(conn *Connect) error {
	// log.Println(c.Front("%v", c.P, msg.Data))

	var err error
	switch msg.Type {
	case 0:
		conn.CloseServer()
	case RTMPTypeSetChunkSize:
		err = msg.solveSetChunkSize(conn)
	case RTMPTypeUserControlMessage:
		err = msg.solveUserControlMessage(conn)
	case RTMPTypeACK:
		err = msg.solveACK(conn)
	case RTMPTypeWindowAcknowledgementSize:
		err = msg.solveWindowsAcknowledgementSize(conn)
	case RTMPTypeSetPeerBandwidth:
		err = msg.solveSetPeerBandwidth(conn)
	case RTMPTypeAudioData:
		err = msg.solveAudioData(conn)
	case RTMPTypeVideoData:
		err = msg.solveVideoData(conn)
	case RTMPTypeAMFData:
		err = msg.solveAMFData(conn)
	case RTMPTypeAMF0Command:
		err = msg.solveAMFCommand(conn, 0)
	case RTMPTypeAMF3Command:
		err = msg.solveAMFCommand(conn, 3)
	default:
		log.Println(c.Front("Unknown message type %d, %v", c.R, msg.Type, msg))
		// err = errors.WithStack(errors.Errorf("Unknown message type %d, %v", msg.Type, msg))
	}
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// solveSetChunkSize 处理 设置分块大小
func (msg *Message) solveSetChunkSize(conn *Connect) error {
	size := lib.ToUint32(msg.Data)
	conn.RecvChunkSize = size
	log.Println(c.Front("Set Recive Chunk Size %d", c.G, size))
	return nil
}

// solveUserControlMessage 处理 用户控制消息
func (msg *Message) solveUserControlMessage(conn *Connect) error {
	size := lib.ToUint32(msg.Data[0:2])
	v2 := lib.ToUint32(msg.Data[2:10])
	log.Println(c.Front("Set Recive Window Acknowledge Size %d", c.G, size, v2))
	return nil
}

// solveACK 处理 确认消息
func (msg *Message) solveACK(conn *Connect) error {
	log.Println(c.Front("ACK", c.G))
	return nil
}

// solveWindowsAcknowledgementSize 处理 设置窗口大小
func (msg *Message) solveWindowsAcknowledgementSize(conn *Connect) error {
	size := lib.ToUint32(msg.Data)
	conn.RecvWindowAcknowledgementSize = size
	log.Println(c.Front("Set Recive Window Acknowledge Size %d", c.G, size))
	return nil
}

// solveSetPeerBandwidth 处理 设置带宽
func (msg *Message) solveSetPeerBandwidth(conn *Connect) error {
	size := lib.ToUint32(msg.Data[0:4])
	bandwidthType := lib.ToUint32(msg.Data[4:5])
	conn.RecvBandwidth = size
	conn.RecvBandwidthType = bandwidthType
	log.Println(c.Front("Set Bandwidth %d %d", c.G, size, bandwidthType))
	return nil
}

// solveAudioData 处理 音频数据
func (msg *Message) solveAudioData(conn *Connect) error {
	message, err := MakeMessage(RTMPTypeAudioData, msg.Data, conn.StreamID, msg.ChunkStreamID, msg.Timestamp)
	if err != nil {
		return errors.WithStack(err)
	}
	conn.TotalTime += message.Timestamp
	message.Timestamp = conn.TotalTime
	conn.WithinStream.Broadcase(message)
	// log.Println(c.Front("Audio Data", c.G))
	return nil
}

// solveVideoData 处理 视频数据
func (msg *Message) solveVideoData(conn *Connect) error {
	message, err := MakeMessage(RTMPTypeVideoData, msg.Data, conn.StreamID, msg.ChunkStreamID, msg.Timestamp)
	if err != nil {
		return errors.WithStack(err)
	}
	conn.TotalTime += message.Timestamp
	message.Timestamp = conn.TotalTime
	conn.WithinStream.Broadcase(message)
	// log.Println(c.Front("Video Data", c.G))
	return nil
}

// solveAMFData 处理AMF Data
func (msg *Message) solveAMFData(conn *Connect) error {
	AMFArray, err := amf.ByteToAMFArray(msg.Data)
	if err != nil {
		return errors.WithStack(err)
	}
	AMFArray2 := make([]interface{}, 0)
	for _, item := range AMFArray {
		AMFArray2 = append(AMFArray2, item.Value())
	}
	return nil
}

// solveAMFCommand 处理 AMF命令
func (msg *Message) solveAMFCommand(conn *Connect, AMFType uint32) error {
	if AMFType == 3 {
		msg.Data = msg.Data[1:]
	}
	AMFArray, err := amf.ByteToAMFArray(msg.Data)
	if err != nil {
		return errors.WithStack(err)
	}
	commandName, ok1 := AMFArray[0].Value().(string)
	transactionID, ok2 := AMFArray[1].Value().(float64)
	commandObject := AMFArray[2].Value()
	var optionalUserArguments interface{}
	if len(AMFArray) >= 4 {
		optionalUserArguments = AMFArray[3].Value()
	}
	if !(ok1 && ok2) {
		return errors.WithStack(errors.New("AMf Command format error"))
	}

	amfCommand := AMFCommand{
		CommandName:           commandName,
		TransactionID:         transactionID,
		CommandObject:         commandObject,
		OptionalUserArguments: optionalUserArguments,
	}

	switch amfCommand.CommandName {
	case "connect":
		msg.solveConnect(conn, &amfCommand)
	case "releaseStream":
		msg.solveReleaseStream(conn, &amfCommand)
	case "FCPublish":
		msg.solveFCPublish(conn, &amfCommand)
	case "createStream":
		msg.solveCreateStream(conn, &amfCommand)
	case "publish":
		msg.solvePublish(conn, &amfCommand)
	case "FCUnpublish":
		msg.solveFCUnpublish(conn, &amfCommand)
	case "getStreamLength":
		msg.solveGetStreamLength(conn, &amfCommand)
	case "play":
		msg.solvePlay(conn, &amfCommand)
	case "FCSubscribe":
		msg.solveFCSubscribe(conn, &amfCommand)
	case "deleteStream":
		msg.solveDeleteStream(conn, &amfCommand)
	default:
		log.Println(c.Front("Unknown AMf command name %s", c.R, amfCommand.CommandName))
		return nil
		// return errors.WithStack(errors.Errorf("Unknown AMf command name %s", amfCommand.CommandName))
	}

	return nil
}

// solveConnect 处理 connect命令
func (msg *Message) solveConnect(conn *Connect, amfCommand *AMFCommand) error {
	commandObject, ok := amfCommand.CommandObject.(map[string]interface{})
	if !ok {
		return errors.WithStack(errors.Errorf("RTMP connect 格式错误"))
	}
	appName, ok := commandObject["app"].(string)
	if !ok {
		return errors.WithStack(errors.Errorf("RTMP connect 格式错误"))
	}
	nameLength := len(appName)
	if appName[nameLength-1] == '/' {
		conn.AppName = appName[:nameLength-1]
	} else {
		conn.AppName = appName
	}

	log.Println(c.Front("connect %v", c.G, amfCommand))

	err := conn.SendWinACKSize(524288)
	if err != nil {
		return errors.WithStack(err)
	}
	err = conn.SendSetPeerBandwidth(524288)
	if err != nil {
		return errors.WithStack(err)
	}
	err = conn.SendStreamBegin(0)
	if err != nil {
		return errors.WithStack(err)
	}
	var objectEncoding float64
	objectEncodingInterface, ok := commandObject["objectEncoding"]
	if ok {
		objectEncoding, _ = objectEncodingInterface.(float64)
	} else {
		objectEncoding = 0
	}

	err = conn.SendResponse(AMFCommand{
		"_result",
		amfCommand.TransactionID,
		map[string]interface{}{
			"mode":         1,
			"capabilities": 31,
			"Author":       "Donview",
			"fmsVer":       "Donview/1.0",
		},
		map[string]interface{}{
			"level":          "status",
			"code":           "NetConnection.Connect.Success",
			"objectEncoding": objectEncoding,
		},
	}, 0, msg.ChunkStreamID)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// solveReleaseStream 处理 releaseStream命令
func (msg *Message) solveReleaseStream(conn *Connect, amfCommand *AMFCommand) error {
	streamName, ok := amfCommand.OptionalUserArguments.(string)
	if !ok {
		return errors.WithStack(errors.Errorf("RTMP releaseStream 格式错误"))
	}

	log.Println(c.Front("releaseStram(%s) %v", c.G, streamName, amfCommand))

	fullName := fmt.Sprintf("%s/%s", conn.AppName, streamName)

	conn.WithinServer.GetStream(fullName).CloseAll()

	return nil
}

// solveFCPublish 处理 FCPublish命令
func (msg *Message) solveFCPublish(conn *Connect, amfCommand *AMFCommand) error {
	streamName, ok := amfCommand.OptionalUserArguments.(string)
	if !ok {
		return errors.WithStack(errors.Errorf("RTMP FCPublish 格式错误"))
	}

	log.Println(c.Front("FCPublish(%s) %v", c.G, streamName, amfCommand))

	conn.StreamName = streamName
	conn.FullName = fmt.Sprintf("%s/%s", conn.AppName, streamName)

	conn.WithinStream = conn.WithinServer.GetStream(conn.FullName)
	conn.WithinStream.AddPublisher(conn)

	return nil
}

// solveCreateStream 处理 createStream命令
func (msg *Message) solveCreateStream(conn *Connect, amfCommand *AMFCommand) error {
	log.Println(c.Front("createStream() %v", c.G, amfCommand))

	streamID := uint32(0)
	if conn.Test {
		streamID = conn.StreamID
	} else {
		conn.Test = true
		streamID = conn.StreamID - 1
	}

	err := conn.SendResponse(AMFCommand{
		"_result",
		amfCommand.TransactionID,
		nil,
		streamID,
	}, 0, 3)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// solvePublish 处理 publish命令
func (msg *Message) solvePublish(conn *Connect, amfCommand *AMFCommand) error {
	log.Println(c.Front("createStream() %v", c.G, amfCommand))

	err := conn.SendResponse(AMFCommand{
		"onStatus",
		0,
		nil,
		map[string]interface{}{
			"level":    "status",
			"clientid": 1,
			"code":     "NetStream.Publish.Start",
		},
	}, msg.StreamID, msg.ChunkStreamID)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// solveUnpublish 处理 FCUnpublish命令
func (msg *Message) solveFCUnpublish(conn *Connect, amfCommand *AMFCommand) error {

	streamName, ok := amfCommand.OptionalUserArguments.(string)
	if !ok {
		return errors.WithStack(errors.Errorf("RTMP FCUnpublish 格式错误"))
	}

	log.Println(c.Front("FCUnpublish(%s) %v", c.G, streamName, amfCommand))

	conn.CloseServer()
	return nil
}

// solveGetStreamLength 处理 getStreamLength命令
func (msg *Message) solveGetStreamLength(conn *Connect, amfCommand *AMFCommand) error {

	streamName, ok := amfCommand.OptionalUserArguments.(string)
	if !ok {
		return errors.WithStack(errors.Errorf("RTMP getStreamLength 格式错误"))
	}

	log.Println(c.Front("getStreamLength(%s) %v", c.G, streamName, amfCommand))

	return nil
}

// solvePlay 处理 play命令
func (msg *Message) solvePlay(conn *Connect, amfCommand *AMFCommand) error {

	streamName, ok := amfCommand.OptionalUserArguments.(string)
	if !ok {
		return errors.WithStack(errors.Errorf("RTMP getStreamLength 格式错误"))
	}

	log.Println(c.Front("play(%s) %v", c.G, streamName, amfCommand))

	conn.StreamName = streamName
	conn.FullName = fmt.Sprintf("%s/%s", conn.AppName, streamName)

	err := conn.SendSetChunkSize(512)
	if err != nil {
		return errors.WithStack(err)
	}

	err = conn.SendStreamIsRecord(conn.StreamID)
	if err != nil {
		return errors.WithStack(err)
	}

	err = conn.SendStreamBegin(conn.StreamID)
	if err != nil {
		return errors.WithStack(err)
	}

	err = conn.SendResponse(AMFCommand{
		"onStatus",
		0,
		nil,
		map[string]interface{}{
			"clientid": 1,
			"code":     "NetStream.Play.Reset",
			"level":    "status",
		},
	}, conn.StreamID, 3)

	if err != nil {
		return errors.WithStack(err)
	}

	err = conn.SendResponse(AMFCommand{
		"onStatus",
		0,
		nil,
		map[string]interface{}{
			"clientid": 1,
			"code":     "NetStream.Play.Start",
			"level":    "status",
		},
	}, conn.StreamID, 3)
	if err != nil {
		return errors.WithStack(err)
	}

	stream := conn.WithinServer.GetStream(conn.FullName)
	conn.WithinStream = stream

	stream.AddReceiver(conn)

	return nil
}

// solveFCSubscribe 处理 FCSubscribe命令
func (msg *Message) solveFCSubscribe(conn *Connect, amfCommand *AMFCommand) error {

	streamName, ok := amfCommand.OptionalUserArguments.(string)
	if !ok {
		return errors.WithStack(errors.Errorf("RTMP FCSubscribe 格式错误"))
	}

	log.Println(c.Front("FCSubscribe(%s) %v", c.G, streamName, amfCommand))

	return nil
}

// solveDeleteStream 处理 deleteStream命令
func (msg *Message) solveDeleteStream(conn *Connect, amfCommand *AMFCommand) error {
	log.Println(c.Front("deleteStream() %v", c.G, amfCommand))

	// stream := conn.WithinServer.GetStream(conn.FullName)
	// stream.DelConnect(conn)
	// conn.CloseServer()

	return nil
}
