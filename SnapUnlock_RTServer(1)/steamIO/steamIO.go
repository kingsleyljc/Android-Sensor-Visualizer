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
	 "os"
	 "io"
	 "strconv"
	 "time"
	 "encoding/csv"
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
 
 type FileBuffer struct{
	 AccelerometerSignal  chan [3]float32
	 SoundSignal chan []int
 }
 
 func InitSteamBuffer() *SteamBuffer {
	 var steamBuffer SteamBuffer
 
	 steamBuffer.AccelerometerSignal = make(chan [3]float32, 100)
	 steamBuffer.SoundSignal = make(chan []int, 100)
 
	 fileBuffer := InitFileBuffer()
 
	 // 防止chan溢出
	 go ReleaseExceededBuffer(&steamBuffer,fileBuffer)
	 return &steamBuffer
 }
 
 func InitFileBuffer() *FileBuffer {
	 var fileBuffer FileBuffer
	 fileBuffer.AccelerometerSignal = make(chan [3]float32, 100)
	 fileBuffer.SoundSignal = make(chan []int, 100)
	 return &fileBuffer
 }
 
 
 var soundSignal = make([]int, 1920)
 var sensorType int
 var Start_record bool
 
 
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
 
 
 func save_acc(fileBuffer *FileBuffer,filename string){
	for{
		if len(fileBuffer.AccelerometerSignal)!=0{ //flag为true才输出
			signal := <- fileBuffer.AccelerometerSignal
			nfs, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
			err = err
			defer nfs.Close()
			nfs.Seek(0, io.SeekEnd)
			w := csv.NewWriter(nfs)
			//设置属性
			w.Comma = ','
			w.UseCRLF = true
			row := []string{strconv.FormatFloat(float64(signal[0]),'f',-5,32),strconv.FormatFloat(float64(signal[1]),'f',-5,32),strconv.FormatFloat(float64(signal[2]),'f',-5,32)}
			w.Write(row)
			w.Flush()
		}
	}
}

func save_audio(fileBuffer *FileBuffer,filename string){
	for{
		if len(fileBuffer.SoundSignal)!=0{ //flag为true才输出
			signal := <- fileBuffer.SoundSignal
			nfs, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
			err = err
			defer nfs.Close()
			nfs.Seek(0, io.SeekEnd)
			w := csv.NewWriter(nfs)
			//设置属性
			w.Comma = ','
			w.UseCRLF = true
			row := []string{strconv.FormatFloat(float64(signal[0]),'f',-5,32),strconv.FormatFloat(float64(signal[1]),'f',-5,32),strconv.FormatFloat(float64(signal[2]),'f',-5,32)}
			w.Write(row)
			w.Flush()
		}
	}
}
 func ReleaseExceededBuffer(steamBuffer *SteamBuffer, fileBuffer *FileBuffer) {
	new_round_acc := true
	new_round_audio := true
	 for {
		// 加速度计buffer >80
		 if len(steamBuffer.AccelerometerSignal) > 80 {
			 AccelerometerSignal := <-steamBuffer.AccelerometerSignal // 改了一下
 
			 if Start_record == true{ // 正在录制
				 fileBuffer.AccelerometerSignal  <- AccelerometerSignal
				 if new_round_acc == true{
					new_round_acc = false
					ti:=time.Now().Nanosecond()
					newFileName  := "./"+strconv.Itoa(ti)+"_acc.csv"
					go save_acc(fileBuffer,newFileName)
				 }
			 } else{
				new_round_acc = true
			 }
 
		 }
 		// 音频buffer >80
		 if len(steamBuffer.SoundSignal) > 90 {
			 for i := 0; i <= 20; i++ {
				 soundSignal := <-steamBuffer.SoundSignal
				 if Start_record == true{ // 正在录制
					fileBuffer.SoundSignal  <- soundSignal
					if new_round_audio == true{
					new_round_audio = false
					   ti:=time.Now().Nanosecond()
					   newFileName  := "./"+strconv.Itoa(ti)+"_audio.csv"
					   go save_audio(fileBuffer,newFileName)
					}
				} else{
				   new_round_audio = true
				}
			 }
		 }
	 }
 }
  