package amf

import (
	"bytes"
	"encoding/binary"
)

// Boolean Boolean类型
type Boolean struct {
	Base
	value bool
}

// NewBooleanDefault 实例化一个Boolean类型
func NewBooleanDefault() AMF {
	return &Boolean{Base{AMFTypeBoolean}, false}
}

// NewBoolean 实例化一个Boolean类型
func NewBoolean(boolean bool) AMF {
	return &Boolean{Base{AMFTypeBoolean}, boolean}
}

// New 从字节流读入AMF Boolean
func (boolean *Boolean) Read(data []byte) error {
	boolean.value = (data[0] != 0)
	return nil
}

// Value 返回对应的Boolean数据
func (boolean *Boolean) Value() interface{} {
	return boolean.value
}

// Length 返回该数据相对于字节流的长度
func (boolean *Boolean) Length() uint32 {
	return 1
}

// Type 返回该数据的Type
func (boolean *Boolean) Type() uint32 {
	return boolean.Base.DataType
}

// Bytes 输出该数据的字节流
func (boolean *Boolean) Bytes() []byte {
	buf := new(bytes.Buffer)

	typeBytes := make([]byte, 2)

	binary.BigEndian.PutUint16(typeBytes, uint16(AMFTypeBoolean))

	buf.Write(typeBytes[1:2]) // 类型
	if boolean.value {
		buf.Write([]byte{0x1}) // 值
	} else {
		buf.Write([]byte{0x0}) // 值
	}

	return buf.Bytes()
}
