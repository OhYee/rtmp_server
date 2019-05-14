package amf

import (
	"bytes"
	"encoding/binary"
)

// ObjectEnd ObjectEnd类型
type ObjectEnd struct {
	Base
}

// NewObjectEndDefault 实例化一个ObjectEnd类型
func NewObjectEndDefault() AMF {
	return &ObjectEnd{Base{AMFTypeObjectEnd}}
}

// NewObjectEnd 实例化一个ObjectEnd类型
func NewObjectEnd() AMF {
	return &ObjectEnd{Base{AMFTypeObjectEnd}}
}

// New 从字节流读入AMF ObjectEnd
func (obj *ObjectEnd) Read(data []byte) error {
	return nil
}

// Value 返回对应的ObjectEnd数据
func (obj *ObjectEnd) Value() interface{} {
	return nil
}

// Length 返回该数据相对于字节流的长度
func (obj *ObjectEnd) Length() uint32 {
	return 0
}

// Type 返回该数据的Type
func (obj *ObjectEnd) Type() uint32 {
	return obj.Base.DataType
}

// Bytes 输出该数据的字节流
func (obj *ObjectEnd) Bytes() []byte {
	buf := new(bytes.Buffer)
	bytes := make([]byte, 2)

	binary.BigEndian.PutUint16(bytes, uint16(AMFTypeObjectEnd))

	buf.Write(bytes[1:2]) // 类型
	return buf.Bytes()
}
