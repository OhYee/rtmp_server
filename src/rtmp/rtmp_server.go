package rtmp

import (
	"database/sql"
	"log"
	"sync"

	c "../lib/colorful"
	// sql
	_ "github.com/go-sql-driver/mysql"
)

// Server RTMP服务
type Server struct {
	db        *sql.DB
	streamMap map[string]*Stream
	mutex     *sync.Mutex
}

// NewServer 新建一个服务
func NewServer(db *sql.DB) Server {
	return Server{
		db:        db,
		streamMap: map[string]*Stream{},
		mutex:     &sync.Mutex{},
	}
}

// GetStream 有一条新的流被建立，获取对应的流信息
func (server *Server) GetStream(streamName string) *Stream {
	defer server.mutex.Unlock()
	server.mutex.Lock()

	stream, ok := server.streamMap[streamName]
	if !ok {
		stream = NewStream(streamName)
		 server.streamMap[streamName]= stream
	}

	return stream
}

// ExecSQL 执行SQL语句
func (server *Server) ExecSQL(sqlStr string, args ...interface{}) {
	defer server.mutex.Unlock()
	server.mutex.Lock()

	stmt, er := server.db.Prepare(sqlStr)
	if er != nil {
		log.Println(c.Front("%s", c.R, er))
	}
	_, er = stmt.Exec(args...)
	if er != nil {
		log.Println(c.Front("%s", c.R, er))
	}
}
