package amf

import (
	"bytes"
	"encoding/binary"
	"math"

	"../../lib"
)

// Number number类型
type Number struct {
	Base
	value float64
}

// NewNumberDefault 实例化一个Number类型
func NewNumberDefault() AMF {
	return &Number{Base{AMFTypeNumber}, 0}
}

// NewNumber 实例化一个Number类型
func NewNumber(number float64) AMF {
	return &Number{Base{AMFTypeNumber}, number}
}

// New 从字节流读入AMF Number
func (number *Number) Read(data []byte) error {
	number.value = lib.ByteToFloat64(data[0:8])
	return nil
}

// Value 返回对应的Number数据
func (number *Number) Value() interface{} {
	return number.value
}

// Length 返回该数据相对于字节流的长度
func (number *Number) Length() uint32 {
	return 8
}

// Type 返回该数据的Type
func (number *Number) Type() uint32 {
	return number.Base.DataType
}

// Bytes 输出该数据的字节流
func (number *Number) Bytes() []byte {
	buf := new(bytes.Buffer)

	typeBytes := make([]byte, 2)
	valueBytes := make([]byte, 8)

	binary.BigEndian.PutUint16(typeBytes, uint16(AMFTypeNumber))
	binary.BigEndian.PutUint64(valueBytes, math.Float64bits(number.value))

	buf.Write(typeBytes[1:2]) // 类型
	buf.Write(valueBytes)     // 值

	return buf.Bytes()
}
