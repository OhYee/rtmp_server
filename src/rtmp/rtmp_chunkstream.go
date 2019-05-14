package rtmp

// ChunkStream 分块流
type ChunkStream struct {
	ChunkStreamID int
}

// NewChunkStream 构造一个分块流
func NewChunkStream(csid int) ChunkStream {
	return ChunkStream{csid}
}

// NewData 该分块流收到的新的分块数据
func (cs *ChunkStream) NewData(data []byte) bool {

	return false
}

// LatestMessage 获取该分块流的最新的一条数据
func (cs *ChunkStream) LatestMessage() Message {
	return Message{}
}
