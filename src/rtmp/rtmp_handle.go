package rtmp

import (
	"log"

	c "../lib/colorful"
	s "../server"
	"github.com/pkg/errors"
)

// HandleConnection RTMP处理函数
func HandleConnection(conn *s.Connect, args interface{}) error {
	defer conn.Close()
	log.Println(c.Front("connect %v", c.G, conn))

	server, ok := args.(*Server)
	if ok {
		connect := NewConnect(conn, server)
		err := connect.Server()
		connect.BeforeClose()
		log.Println(c.Front("disconnect %v", c.R, conn))
		if err != nil {
			log.Println(c.Front("Connect: %v", c.Y, connect))
			log.Println(c.Front("Error: %+v", c.R, err))
			return errors.WithStack(err)
		}
	} else {
		log.Println(c.Front("Error: RTMP server error.", c.R))
	}

	return nil
}
