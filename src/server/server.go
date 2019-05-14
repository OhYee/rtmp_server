package server

import (
	"fmt"
	"log"
	"net"

	"github.com/pkg/errors"
)

// Server ...
type Server struct {
	Protocol string
	Address  string
	Port     int
	Handle   func(*Connect, interface{}) error
	Args     interface{}
}

// Listen 监听某个端口
func (s Server) Listen() error {
	log.Printf("Server start at http://%s:%d (%s).\n", s.Address, s.Port, s.Protocol)
	ln, err := net.Listen(s.Protocol, fmt.Sprintf(":%d", s.Port))
	if err != nil {
		log.Println(err)
		return errors.WithStack(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go s.Handle(&Connect{conn}, s.Args)
	}
}
