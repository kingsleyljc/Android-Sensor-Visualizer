/*
 * @Author: Neotter
 * @Date: 2020-09-02 12:58:12
 * @LastEditTime: 2020-09-12 17:39:28
 * @LastEditors: Please set LastEditors
 * @Description: In User Settings Edit
 * @FilePath: \GoProject\SnapUnlock_RTServer\util\ByteUtil.go
 */
package util

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

// 定义Byte Array的顺序是大端还是小端
var LittleEndian endian = true

var BigEndian endian = false

type endian = bool

//Byte数组转成int(有符号)
func Bytes2Int(b []byte, e endian) (int, error) {
	if len(b) == 3 {
		b = append([]byte{0}, b...)
	}
	bytesBuffer := bytes.NewBuffer(b)
	var order binary.ByteOrder
	switch e {
	case LittleEndian:
		order = binary.LittleEndian
	case BigEndian:
		order = binary.BigEndian
	}

	switch len(b) {
	case 1:
		var tmp int8
		err := binary.Read(bytesBuffer, order, &tmp)
		return int(tmp), err
	case 2:
		var tmp int16
		err := binary.Read(bytesBuffer, order, &tmp)
		return int(tmp), err
	case 4:
		var tmp int32
		err := binary.Read(bytesBuffer, order, &tmp)
		return int(tmp), err
	default:
		return 0, fmt.Errorf("%s", "BytesToInt bytes lenth is invaild!")
	}
}

//Byte数组转成float32(有符号)
func Byte2Float32(b []byte, e endian) float32 {
	var bits uint32
	switch e {
	case LittleEndian:
		bits = binary.LittleEndian.Uint32(b)
	case BigEndian:
		bits = binary.BigEndian.Uint32(b)
	}
	return math.Float32frombits(bits)
}

//Byte数组转成float64(有符号)
func Byte2Float64(b []byte, e endian) float64 {
	var bits uint64
	switch e {
	case LittleEndian:
		bits = binary.LittleEndian.Uint64(b)
	case BigEndian:
		bits = binary.BigEndian.Uint64(b)
	}
	return math.Float64frombits(bits)
}

func Float32ToByte(f float32, e endian) []byte {
	var buf [4]byte
	switch e {
	case LittleEndian:
		binary.LittleEndian.PutUint32(buf[:], math.Float32bits(f))
	case BigEndian:
		binary.BigEndian.PutUint32(buf[:], math.Float32bits(f))
	}
	return buf[:]
}

func Float64ToByte(f float64, e endian) []byte {
	var buf [8]byte
	switch e {
	case LittleEndian:
		binary.LittleEndian.PutUint64(buf[:], math.Float64bits(f))
	case BigEndian:
		binary.BigEndian.PutUint64(buf[:], math.Float64bits(f))
	}
	return buf[:]
}
