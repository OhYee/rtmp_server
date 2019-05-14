package server

import "net"

// Connect 对于net中Conn的加强
type Connect struct {
	net.Conn
}

// ReadLength 读入指定长度的数据
func (conn Connect) ReadLength(len int) ([]byte, error) {
	data := make([]byte, len)
	_, err := conn.Read(data)
	if err != nil {
		return data, nil
	}

	return data, nil
}
