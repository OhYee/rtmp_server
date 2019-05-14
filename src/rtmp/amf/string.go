package amf

import (
	"bytes"
	"encoding/binary"

	"../../lib"
)

// String string类型
type String struct {
	Base
	value string
	len   uint32
}

// NewStringDefault 实例化一个String类型
func NewStringDefault() AMF {
	return &String{Base{AMFTypeString}, "", 0}
}

// NewString 实例化一个String类型
func NewString(str string) AMF {
	return &String{Base{AMFTypeString}, str, uint32(len(str))}
}

// Read 从字节流读入AMF String
func (str *String) Read(data []byte) error {
	str.len = lib.ToUint32(data[0:2])
	str.value = string(data[2 : 2+str.len])
	return nil
}

// Value 返回对应的String数据
func (str *String) Value() interface{} {
	return str.value
}

// Length 返回该数据相对于字节流的长度
func (str *String) Length() uint32 {
	return str.len + 2
}

// Type 返回该数据的Type
func (str *String) Type() uint32 {
	return str.Base.DataType
}

// Bytes 输出该数据的字节流
func (str *String) Bytes() []byte {
	buf := new(bytes.Buffer)

	typeBytes := make([]byte, 2)
	lenBytes := make([]byte, 2)

	binary.BigEndian.PutUint16(typeBytes, uint16(AMFTypeString))
	binary.BigEndian.PutUint16(lenBytes, uint16(str.len))

	buf.Write(typeBytes[1:2])    // 类型
	buf.Write(lenBytes[0:2])     // 长度
	buf.Write([]byte(str.value)) // 长度字节内容

	return buf.Bytes()
}
