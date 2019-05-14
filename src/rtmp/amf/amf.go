package amf

import (
	"github.com/pkg/errors"
)

// AMF 类型常量
const (
	AMFTypeNumber    = uint32(0x00)
	AMFTypeBoolean   = uint32(0x01)
	AMFTypeString    = uint32(0x02)
	AMFTypeObject    = uint32(0x03)
	AMFTypeNone      = uint32(0x05)
	AMFTypeECMAArray = uint32(0x08)
	AMFTypeObjectEnd = uint32(0x09)
)

// Base 基类
type Base struct {
	DataType uint32
}

// AMF AMF结构的值
type AMF interface {
	Read([]byte) error
	Value() interface{}
	Length() uint32
	Type() uint32
	Bytes() []byte
}

// Pair Key-Value二元组
type Pair struct {
	key   string
	value AMF
}

// NewAMF 构造AMF对象
func NewAMF(data []byte) (AMF, error) {
	typeID := uint32(data[0])
	var value AMF
	switch typeID {
	case AMFTypeNumber:
		value = NewNumberDefault()
	case AMFTypeBoolean:
		value = NewBooleanDefault()
	case AMFTypeString:
		value = NewStringDefault()
	case AMFTypeObject:
		value = NewObjectDefault()
	case AMFTypeECMAArray:
		value = NewECMAArrayDefault()
	case AMFTypeObjectEnd:
		value = NewObjectEndDefault()
	case AMFTypeNone:
		value = NewNone()
	default:
		// log.Println(data)
		return value, errors.WithStack(errors.Errorf("RTMP message AMF type error %d", typeID))
	}
	err := value.Read(data[1:])
	if err != nil {
		return value, errors.WithStack(err)
	}
	// log.Println(value, data)
	return value, nil
}

// ByteToAMFArray 将字节流转换为AMF对象列表
func ByteToAMFArray(data []byte) ([]AMF, error) {
	var array []AMF

	for len(data) > 0 {
		amf, err := NewAMF(data)
		if err != nil {
			return array, errors.WithStack(err)
		}
		array = append(array, amf)
		data = data[amf.Length()+1:]
		// log.Printf("%+v\n%+v\n", amf, data)
	}
	return array, nil
}

// MakeAMF 将一个Go类型变量转换为AMF对象
func MakeAMF(data interface{}) (AMF, error) {
	var err error
	amf := NewNone()

	switch data.(type) {
	case bool:
		amf = NewBoolean(data.(bool))
	case nil:
		amf = NewNone()
	case float64:
		amf = NewNumber(data.(float64))
	case int:
		amf = NewNumber(float64(data.(int)))
	case uint32:
		amf = NewNumber(float64(data.(uint32)))
	case map[string]interface{}:
		amf, err = NewObject(data.(map[string]interface{}))
	case string:
		amf = NewString(data.(string))
	default:
		amf = NewNone()
	}
	return amf, err
}
