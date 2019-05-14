package rtmp

import (
	"log"
	"sync"

	c "../lib/colorful"
)

/*

	对应一个RTMP流数据 如 myapp/mystream

*/

// Stream RTMP流，有一个输入流id，多个输出流id
type Stream struct {
	Name      string      // 流名称
	Publisher *Connect    // 输入流
	Receivers []*Connect  // 输出流
	mutex     *sync.Mutex //锁
	Tag       *Message    // 视频Tag
}

// NewStream 新建一个流
func NewStream(fullName string) *Stream {
	stream := Stream{
		Name:      fullName,
		Publisher: nil,
		Receivers: make([]*Connect, 0),
		mutex:     &sync.Mutex{},
		Tag:       nil,
	}
	return &stream
}

// GetTag 获取音视频标签
func (stream *Stream) GetTag() (Message, bool) {
	// defer stream.mutex.Unlock()
	// stream.mutex.Lock()

	if stream.Tag == nil {
		return Message{}, false
	}
	return stream.Tag.Copy(), true
}

// Broadcase 在当前流内广播对应数据
func (stream *Stream) Broadcase(data Message) {
	defer stream.mutex.Unlock()
	stream.mutex.Lock()

	data = data.Copy()

	if stream.Tag == nil && data.Type == RTMPTypeVideoData {
		tag := data.Copy()
		stream.Tag = &tag
	} else {
		// 当有新的音视频消息时，应该先将其存入StreamData
		// stream.StreamData = append(stream.StreamData, data)

		for _, conn := range stream.Receivers {
			// log.Println("chan", len(conn.Thread.MessageChannel))
			conn.Thread.MessageChannel <- data
		}
	}

}

// AddPublisher 在当前流中增加一个推流端
func (stream *Stream) AddPublisher(conn *Connect) {
	defer stream.mutex.Unlock()
	stream.mutex.Lock()

	stream.Publisher = conn
	conn.WithinServer.ExecSQL("REPLACE INTO connect(`url`,`datetime`) VALUES(?,NOW())", conn.FullName)
}

// AddReceiver 在当前流中增加一个拉流端
func (stream *Stream) AddReceiver(conn *Connect) {
	defer stream.mutex.Unlock()
	stream.mutex.Lock()

	log.Println(c.Front("AddReceiver %s", c.G, stream.Name))
	conn.Thread.Start()

	stream.Receivers = append(stream.Receivers, conn)

	// for _, msg := range stream.StreamData {
	// 	// conn.Chan <- msg
	// 	conn.SendAVMessage(msg)
	// }

}

// DelConnect 在当前流删除连接，自动判断属于推流端还是拉流端
func (stream *Stream) DelConnect(conn *Connect) {
	defer stream.mutex.Unlock()
	stream.mutex.Lock()

	if conn == stream.Publisher {
		conn.WithinServer.ExecSQL("DELETE from connect WHERE `url`=?", conn.FullName)
		stream.closeAll()
	} else {
		stream.delReceiver(conn)
	}
}

// delReceiver 在当前流中删除一个拉流端
func (stream *Stream) delReceiver(conn *Connect) {
	index := -1
	for idx, _conn := range stream.Receivers {
		if conn == _conn {
			index = idx
			break
		}
	}
	if index != -1 {
		stream.Receivers = append(stream.Receivers[:index], stream.Receivers[index+1:]...)
	}
}

// CloseAll 断开该流的所有连接
func (stream *Stream) CloseAll() {
	defer stream.mutex.Unlock()
	stream.mutex.Lock()

	stream.closeAll()
}

// closeAll 断开该流的所有连接
func (stream *Stream) closeAll() {
	if stream.Publisher != nil {
		stream.Publisher.CloseServer()
	}
	for _, conn := range stream.Receivers {
		conn.CloseServer()
	}
	stream.Publisher = nil
	stream.Receivers = stream.Receivers[0:0]
}
