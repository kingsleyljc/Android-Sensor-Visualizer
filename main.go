/*
 * @Author: your name
 * @Date: 2020-08-28 21:35:41
 * @LastEditTime: 2020-09-25 21:47:07
 * @LastEditors: Please set LastEditors
 * @Description: In User Settings Edit
 * @FilePath: \GoProject\SnapUnlock_RTServer\hello.go
 */
package main

import (
	"SnapUnlock_RTServer/steamIO"
	"SnapUnlock_RTServer/util"

	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/r9y9/gossp/dct"
)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var steamBuffer *steamIO.SteamBuffer

func receiver(c *gin.Context) {
	//升级get请求为webSocket协议
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()
	defer fmt.Println("receiver.Close")
	//// 以上可以不看
	fmt.Println("receive.Open")
	soundSignal := make([]float64, 1920)
	soundMessage := make([]byte, 7685)

	for {
		//读取ws中的数据
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}
		steamIO.Write2Buffer(&message, steamBuffer)
		if len(message) == 17 {
			// 不断地发送到manager需要broadcast的channel中
			manager.broadcast <- message
		} else {
			// 把byte转成int再转成float64
			for i, j := 5, 0; i < len(message); i = i + 2 {
				resInt, _ := util.Bytes2Int(message[i:i+2], util.LittleEndian)
				soundSignal[j] = float64(resInt)
				j++
			}
			// dct变换
			soundSignal = dct.DCT(soundSignal)
			// float64再转成float32再转成bytes
			soundMessage[0] = message[0]
			soundMessage[1] = message[1]
			soundMessage[2] = message[2]
			soundMessage[3] = message[3]
			soundMessage[4] = message[4]
			length := len(soundSignal)
			for i, j := 0, 5; i < length; i++ {
				res := util.Float32ToByte(float32(soundSignal[i]), util.LittleEndian)
				soundMessage[j] = res[0]
				j++
				soundMessage[j] = res[1]
				j++
				soundMessage[j] = res[2]
				j++
				soundMessage[j] = res[3]
				j++
			}
			// 发送
			manager.broadcast <- soundMessage

		}
	}
}

// 广播消息到客户端

type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

type Client struct {
	socket *websocket.Conn
	send   chan []byte
}

type Message struct {
	Content string `json:"content,omitempty"`
}

// 实现ClientManager
var manager = ClientManager{
	broadcast:  make(chan []byte, 100),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

// 轮询放入注册，取消注册和message数据
func (manager *ClientManager) start() {
	//go (*manager).releaseExceededBuffer()
	for {
		select {
		case conn := <-manager.register:
			manager.clients[conn] = true
		case conn := <-manager.unregister:
			if _, ok := manager.clients[conn]; ok {
				close(conn.send)
				delete(manager.clients, conn)
			}
		case message := <-manager.broadcast:
			// 一旦broadcast里面有消息就塞到manager.clients中的send中
			for conn := range manager.clients {
				select {
				case conn.send <- message:
				}
			}
		}
	}
}

// 放在每一个接上服务器的连接中运行，用于不断地发送数据
func (c *Client) send2Client() {
	defer func() {
		c.socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				fmt.Println("send2Client() Error!")
			}
			c.socket.WriteMessage(websocket.BinaryMessage, message)
		}

	}
}

// 广播数据
func wsPage(c *gin.Context) {
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		return
	}
	//defer ws.Close()
	fmt.Println("WebSocket Open")

	client := &Client{socket: ws, send: make(chan []byte)}
	// 注册client到manager
	manager.register <- client
	go client.send2Client()
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}
		if string(message) == "1" {
			if steamIO.Start_record == true {
				steamIO.Start_record = false
			} else {
				steamIO.Start_record = true
			}

		}

		if steamIO.Start_record == true {
			fmt.Println("okok")
		} else {
			fmt.Println("nono")
		}

	}

}

/// 发送主页
func index(c *gin.Context) {
	c.HTML(200, "index.html", "kumiko")
}

/// 发送waiting.png
func pic(c *gin.Context) {
	c.File("./html/waiting.png")
}

func main() {
	fmt.Println("Server on ")
	steamBuffer = steamIO.InitSteamBuffer()
	bindAddress := "192.168.1.103:7777"
	go manager.start()
	r := gin.Default()
	//pprof.Register(r)
	r.LoadHTMLFiles("html/index.html")
	r.GET("/receiver", receiver) //接收andorid端的数据， websocket 协议
	r.GET("/", index)
	r.GET("/waiting.png", pic)
	r.GET("/ws", wsPage) // 发送到网页端， websocket 协议
	r.Run(bindAddress)
}
