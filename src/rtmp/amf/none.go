package amf

import (
	"bytes"
	"encoding/binary"
)

// None None类型
type None struct {
	Base
}

// NewNoneDefault 实例化一个None类型
func NewNoneDefault() AMF {
	return &None{Base{AMFTypeNone}}
}

// NewNone 实例化一个None类型
func NewNone() AMF {
	return &None{Base{AMFTypeNone}}
}

// New 从字节流读入AMF None
func (obj *None) Read(data []byte) error {
	return nil
}

// Value 返回对应的None数据
func (obj *None) Value() interface{} {
	return nil
}

// Length 返回该数据相对于字节流的长度
func (obj *None) Length() uint32 {
	return 0
}

// Type 返回该数据的Type
func (obj *None) Type() uint32 {
	return obj.Base.DataType
}

// Bytes 输出该数据的字节流
func (obj *None) Bytes() []byte {
	buf := new(bytes.Buffer)
	bytes := make([]byte, 2)

	binary.BigEndian.PutUint16(bytes, uint16(AMFTypeNone))

	buf.Write(bytes[1:2]) // 类型
	return buf.Bytes()
}
