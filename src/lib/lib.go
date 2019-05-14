package lib

import (
	"encoding/binary"
	"math"
)

// ByteArrayConcat 拼接字节切片
func ByteArrayConcat(arrays ...[]byte) []byte {
	totalLength := 0
	for _, array := range arrays {
		totalLength += len(array)
	}
	byteArray := make([]byte, 0, totalLength)
	for _, array := range arrays {
		byteArray = append(byteArray, array...)
	}
	return byteArray
}

// ToUint32 将一个字节切片转换为对应的int
func ToUint32(byteArray []byte) uint32 {
	var result uint32
	for _, num := range byteArray {
		result = result*256 + uint32(num)
	}
	// log.Println(byteArray, result)

	return result
}

// ByteToFloat64 将字节流转换为浮点数
func ByteToFloat64(byteArray []byte) float64 {
	// log.Println(len(byteArray))
	bits := binary.BigEndian.Uint64(byteArray)
	return math.Float64frombits(bits)
}

// Min 返回最小值
func Min(nums ...uint32) uint32 {
	minNum := nums[0]
	for _, num := range nums {
		if num < minNum {
			minNum = num
		}
	}
	return minNum
}
