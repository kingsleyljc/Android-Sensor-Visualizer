/*
 * @Author: your name
 * @Date: 2020-09-02 21:20:36
 * @LastEditTime: 2020-09-22 15:54:06
 * @LastEditors: Please set LastEditors
 * @Description: In User Settings Edit
 * @FilePath: \GoProject\SnapUnlock_RTServer\steamIO\steamIO.go
 */
package steamIO

import (
	"SnapUnlock_RTServer/sensors"
	"SnapUnlock_RTServer/util"
	"fmt"
)

type SteamBuffer struct {
	AccelerometerSignal chan [3]float32
	// // Axix X Y Z
	// AxisX chan float32
	// AxisY chan float32
	// AxisZ chan float32
	// RawSignal Size = 3840, convert to float64, size of soundSignal is 3840/2 = 1920
	SoundSignal chan []int
}

func InitSteamBuffer() *SteamBuffer {
	var steamBuffer SteamBuffer

	steamBuffer.AccelerometerSignal = make(chan [3]float32, 100)
	steamBuffer.SoundSignal = make(chan []int, 100)
	// 防止chan溢出
	go ReleaseExceededBuffer(&steamBuffer)
	return &steamBuffer
}

var soundSignal = make([]int, 1920)
var sensorType int

func Write2Buffer(message *[]byte, steamBuffer *SteamBuffer) {

	sensorType, _ = util.Bytes2Int((*message)[1:5], util.LittleEndian)

	switch sensorType, _ = util.Bytes2Int((*message)[1:5], util.LittleEndian); sensorType {

	case sensors.ACCELEROMETER: //加速度计
		steamBuffer.AccelerometerSignal <- [3]float32{util.Byte2Float32((*message)[5:9], util.LittleEndian), util.Byte2Float32((*message)[9:13], util.LittleEndian), util.Byte2Float32((*message)[13:17], util.LittleEndian)}

	case sensors.MICROPHONE: // 麦克风
		j := 0
		for i := 5; i < len(*message); i = i + 2 {
			soundSignal[j], _ = util.Bytes2Int((*message)[i:i+2], util.LittleEndian)
			j++
		}
		steamBuffer.SoundSignal <- soundSignal
	default:
		fmt.Println("UnknowSensor")
	}
}

func ReleaseExceededBuffer(steamBuffer *SteamBuffer) {
	for {
		if len(steamBuffer.AccelerometerSignal) > 90 {
			AccelerometerSignal := <-steamBuffer.AccelerometerSignal // 改了一下
			AccelerometerSignal = AccelerometerSignal
			//fmt.Printf("RelesseBuffer_ACC X: %v \n", AccelerometerSignal)
		}
		if len(steamBuffer.SoundSignal) > 90 {
			for i := 0; i <= 20; i++ {
				soundSignal := <-steamBuffer.SoundSignal
				soundSignal = soundSignal
				//fmt.Printf("RelesseBuffer_Sound: %d \n", soundSignal)
			}
		}
	}
}
