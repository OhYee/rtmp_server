package amf

import (
	"bytes"
	"encoding/binary"

	"../../lib"
	"github.com/pkg/errors"
)

// ECMAArray ECMAArray类型
type ECMAArray struct {
	Base
	value  []Pair
	length uint32
	Num    uint32
}

// NewECMAArrayDefault 实例化一个ECMAArray类型
func NewECMAArrayDefault() AMF {
	return &ECMAArray{Base{AMFTypeECMAArray}, []Pair{}, 0, 0}
}

// NewECMAArray 实例化一个ECMAArray类型
func NewECMAArray(m map[string]interface{}) (AMF, error) {
	amf := &ECMAArray{Base{AMFTypeECMAArray}, []Pair{}, 0, uint32(len(m))}
	var err error
	for key, value := range m {
		var amfValue AMF
		switch value.(type) {
		case float64:
			valueFloat := value.(float64)
			amfValue = NewNumber(valueFloat)
		case string:
			valueString := value.(string)
			amfValue = NewString(valueString)
		case map[string]interface{}:
			valueECMAArray := value.(map[string]interface{})
			amfValue, err = NewECMAArray(valueECMAArray)
		}
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

// New 从字节流读入AMF ECMAArray
func (arr *ECMAArray) Read(data []byte) error {
	var err error
	// 读入个数
	arr.length = 0
	arr.Num = lib.ToUint32(data[arr.length+0 : arr.length+4])
	arr.length += 4

	for {
		keyLength := lib.ToUint32(data[arr.length+0 : arr.length+2])
		var pair Pair
		pair.key = string(data[arr.length+2 : arr.length+2+keyLength])
		pair.value, err = NewAMF(data[arr.length+2+keyLength:])
		if err != nil {
			return errors.WithStack(err)
		}
		arr.length += 2 + keyLength + 1 + pair.value.Length()
		arr.value = append(arr.value, pair)
		if pair.value.Type() == AMFTypeObjectEnd {
			break
		}
	}
	return nil
}

// Value 返回对应的ECMAArray数据
func (arr *ECMAArray) Value() interface{} {
	m := make(map[string]interface{})
	for _, item := range arr.value {
		m[item.key] = item.value.Value()
	}
	return m
}

// Length 返回该数据相对于字节流的长度
func (arr *ECMAArray) Length() uint32 {
	return arr.length
}

// Type 返回该数据的Type
func (arr *ECMAArray) Type() uint32 {
	return arr.Base.DataType
}

// Bytes 输出该数据的字节流
func (arr *ECMAArray) Bytes() []byte {
	buf := new(bytes.Buffer)
	bytes := make([]byte, 4)

	binary.BigEndian.PutUint16(bytes, uint16(AMFTypeECMAArray))
	buf.Write(bytes[1:2]) // 类型
	binary.BigEndian.PutUint32(bytes, uint32(arr.Num))
	buf.Write(bytes[0:4]) // 类型
	for _, pair := range arr.value {
		buf.Write(NewString(pair.key).Bytes()[1:])
		buf.Write(pair.value.Bytes())
	}
	return buf.Bytes()
}
