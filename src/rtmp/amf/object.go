package amf

import (
	"bytes"
	"encoding/binary"

	"../../lib"
	"github.com/pkg/errors"
)

// Object Object类型
type Object struct {
	Base
	value  []Pair
	length uint32
}

// NewObjectDefault 实例化一个Object类型
func NewObjectDefault() AMF {
	return &Object{Base{AMFTypeObject}, []Pair{}, 0}
}

// NewObject 实例化一个Object类型
func NewObject(m map[string]interface{}) (AMF, error) {
	amf := &Object{Base{AMFTypeObject}, []Pair{}, 0}
	for key, value := range m {
		amfValue, err := MakeAMF(value)
		if err != nil {
			return amf, errors.WithStack(err)
		}
		amf.length = amf.length + amfValue.Length()
		amf.value = append(amf.value, Pair{key, amfValue})
	}
	amf.length = amf.length + 3
	amf.value = append(amf.value, Pair{"", NewObjectEnd()})
	return amf, nil
}

// New 从字节流读入AMF Object
func (obj *Object) Read(data []byte) error {
	var err error
	obj.length = 0
	for {
		keyLength := lib.ToUint32(data[obj.length+0 : obj.length+2])
		var pair Pair
		pair.key = string(data[obj.length+2 : obj.length+2+keyLength])
		pair.value, err = NewAMF(data[obj.length+2+keyLength:])
		if err != nil {
			return errors.WithStack(err)
		}
		obj.length += 2 + keyLength + 1 + pair.value.Length()
		obj.value = append(obj.value, pair)
		if pair.value.Type() == AMFTypeObjectEnd {
			break
		}
	}
	return nil
}

// Value 返回对应的Object数据
func (obj *Object) Value() interface{} {
	m := make(map[string]interface{})
	for _, item := range obj.value {
		m[item.key] = item.value.Value()
	}
	return m
}

// Length 返回该数据相对于字节流的长度
func (obj *Object) Length() uint32 {
	return obj.length
}

// Type 返回该数据的Type
func (obj *Object) Type() uint32 {
	return obj.Base.DataType
}

// Bytes 输出该数据的字节流
func (obj *Object) Bytes() []byte {
	buf := new(bytes.Buffer)
	bytes := make([]byte, 2)

	binary.BigEndian.PutUint16(bytes, uint16(AMFTypeObject))
	buf.Write(bytes[1:2]) // 类型
	for _, pair := range obj.value {
		buf.Write(NewString(pair.key).Bytes()[1:])
		buf.Write(pair.value.Bytes())
	}
	return buf.Bytes()
}
