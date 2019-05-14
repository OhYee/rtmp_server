package rtmp

import "log"

// AVThread 音视频发送线程
type AVThread struct {
	Conn           *Connect
	MessageChannel chan Message
	CommandChannel chan string
	isWorking      bool
}

// NewAVThread 新建一个音视频发送线程
func NewAVThread(conn *Connect) AVThread {
	return AVThread{
		Conn:           conn,
		MessageChannel: make(chan Message, 8),
		CommandChannel: make(chan string),
		isWorking:      false,
	}
}

// Start 启动音视频线程
func (thread *AVThread) Start() {
	if thread.isWorking {
		log.Println("Thread is already working")
	} else {
		thread.isWorking = true
		go thread.loop()
	}
}

// Stop 结束线程
func (thread *AVThread) Stop() {
	if thread.isWorking {
		thread.isWorking = false
		thread.CommandChannel <- "stop"
	}
}

func (thread *AVThread) loop() {
	for {
		select {
		case msg := <-thread.MessageChannel:
			thread.Conn.SendAVMessage(msg)
		case command := <-thread.CommandChannel:
			if command == "stop" {
				return
			}
		}
	}
}
