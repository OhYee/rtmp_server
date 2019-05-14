package rtmp

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"log"
	"math/rand"

	"../lib"
	c "../lib/colorful"
	"github.com/pkg/errors"
)

// FPKey RTMP协议加密key - 客户端
var FPKey = []byte{
	0x47, 0x65, 0x6E, 0x75, 0x69, 0x6E, 0x65, 0x20,
	0x41, 0x64, 0x6F, 0x62, 0x65, 0x20, 0x46, 0x6C,
	0x61, 0x73, 0x68, 0x20, 0x50, 0x6C, 0x61, 0x79,
	0x65, 0x72, 0x20, 0x30, 0x30, 0x31, // Genuine Adobe Flash Player 001
	0xF0, 0xEE, 0xC1, 0x4A, 0x80, 0x68, 0xBE, 0xE8,
	0x2E, 0x00, 0xD0, 0xD1, 0x02, 0x9E, 0x7E, 0x57,
	0x6E, 0xEC, 0x5D, 0x2D, 0x29, 0x80, 0x6F, 0xAB,
	0x93, 0xB8, 0xE6, 0x36, 0xCF, 0xEB, 0x31, 0xAE,
}

// FMSKey RTMP协议加密Key - 服务端
var FMSKey = []byte{
	0x47, 0x65, 0x6e, 0x75, 0x69, 0x6e, 0x65, 0x20,
	0x41, 0x64, 0x6f, 0x62, 0x65, 0x20, 0x46, 0x6c,
	0x61, 0x73, 0x68, 0x20, 0x4d, 0x65, 0x64, 0x69,
	0x61, 0x20, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x20, 0x30, 0x30, 0x31, // Genuine Adobe Flash Media Server 001
	0xf0, 0xee, 0xc2, 0x4a, 0x80, 0x68, 0xbe, 0xe8,
	0x2e, 0x00, 0xd0, 0xd1, 0x02, 0x9e, 0x7e, 0x57,
	0x6e, 0xec, 0x5d, 0x2d, 0x29, 0x80, 0x6f, 0xab,
	0x93, 0xb8, 0xe6, 0x36, 0xcf, 0xeb, 0x31, 0xae,
} // 68

// Handshake 握手过程处理
func Handshake(conn *Connect) error {
	var S0, S1, S2 []byte

	// C0
	C0, err := conn.Read(1)
	if err != nil {
		return errors.WithStack(err)
	}
	if len(C0) == 0 {
		log.Printf("C0: %v\n", C0)
	}

	// C1
	C1, err := conn.Read(1536)
	if err != nil {
		return errors.WithStack(err)
	}
	// log.Printf("C1: %vn", C1)
	if lib.ToUint32(C1[4:8]) == 0 {
		// 简单握手
		if !simpleHandshake(C1) {
			return errors.WithStack(errors.New("RTMP connect error: Simple handshake vaild error"))
		}

		S0 = []byte{3}
		S1 = C1
		S2 = C1
	} else {
		// 复杂握手
		scheme := 0
		match, digest := complexHandshake(C1, scheme)
		if !match {
			scheme = 1
			match, digest = complexHandshake(C1, scheme)
			if !match {
				return errors.WithStack(errors.New("RTMP connect error: Complex handshake vaild error"))
			}
		}
		log.Println(c.Front("ComplexHandshake key:%v scheme:%d", c.G, digest, scheme))

		// S0 S1 S2
		S0 = []byte{3}

		S1 = append([]byte{0x00, 0x00, 0x00, 0x00, 0x04, 0x05, 0x00, 0x01}, make([]byte, 1536-8)...)
		for i := 8; i < 1536; i++ {
			S1[i] = byte(rand.Intn(256))
		}
		offset := getDigestOffset(S1, scheme)
		digestS1 := HMACSha256(FMSKey[:36], lib.ByteArrayConcat(S1[:offset], S1[offset+32:]))
		S1 = lib.ByteArrayConcat(S1[0:offset], digestS1, S1[offset+32:])

		S2 = make([]byte, 1536-32)
		for i := 0; i < 1536-32; i++ {
			S2[i] = byte(rand.Intn(256))
		}
		digestS2 := HMACSha256(HMACSha256(FMSKey[:68], digest), S2[:1536-32])
		S2 = lib.ByteArrayConcat(S2, digestS2)
	}

	// 发送S0 S1 S2
	conn.Write(S0)
	conn.Write(S1)
	conn.Write(S2)

	// C2
	C2, err := conn.Read(1536)
	if err != nil {
		return errors.WithStack(err)
	}
	if len(C2) != 1536 {
		log.Printf("C2: %v\n", C2)
	}

	return nil
}

// simpleHandshake 简单握手,不做判断
func simpleHandshake(C1 []byte) bool {
	log.Println("Simple Handshake")
	return true
}

// complexHandShake 复杂握手
func complexHandshake(C1 []byte, scheme int) (bool, []byte) {
	log.Printf("Complex Handshake %d\n", scheme)

	offset := getDigestOffset(C1, scheme)

	P1 := C1[:offset]
	digestData := C1[offset : offset+32]
	P2 := C1[offset+32:]

	digest := HMACSha256(FPKey[:30], lib.ByteArrayConcat(P1, P2))

	if bytes.Equal(digest, digestData) {
		return true, digest
	}
	return false, digest
}

// getDigestOffset 获得数据中32字节的摘要值的偏移地址
func getDigestOffset(data []byte, scheme int) int {
	digestPos := 4 + 4
	if scheme == 1 {
		digestPos += 764
	}
	// 这里byte是int8，必须转换为int32，不然会溢出
	offset := (int(data[digestPos+0]) + int(data[digestPos+1]) + int(data[digestPos+2]) + int(data[digestPos+3])) % 728
	beforeKey := digestPos + 4 + offset
	return beforeKey
}

// HMACSha256 使用HMAC Sha256 加密字符串
func HMACSha256(key []byte, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
