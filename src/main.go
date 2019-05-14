package main

import (
	"database/sql"
	"log"

	"./rtmp"

	"./config"
	"./server"
)

func main() {
	db, err := sql.Open("mysql", config.GetDatabaseData())
	if err != nil {
		log.Println(err)
	}
	rtmpServer := rtmp.NewServer(db)

	s := server.Server{
		Protocol: "tcp",
		Address:  "127.0.0.1",
		Port:     19356,
		Handle:   rtmp.HandleConnection,
		Args:     &rtmpServer,
	}
	s.Listen()
}
