package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/pkg/protocol"
	"log"

	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
)

type Client struct {
	Ws   *websocket.Conn
	UUID string
	Send chan []byte
}

func NewClient(name string, ws *websocket.Conn) *Client {
	return &Client{
		Ws:   ws,
		UUID: name,
		Send: make(chan []byte),
	}
}

func (c *Client) Read() {
	defer func() {
		ServerInstance.Ungister <- c
		c.Ws.Close()
	}()
	for {
		c.Ws.PongHandler()
		_, message, err := c.Ws.ReadMessage()
		if err != nil {
			log.Println("client read message error")
			ServerInstance.Ungister <- c
			c.Ws.Close()
			break
		}

		// 将消息编码成protobuf
		pbMessage := protocol.Message{}
		proto.Unmarshal(message, &pbMessage)
		// 心跳消息
		if pbMessage.Type == "heatbeat" {
			pong := &protocol.Message{
				Type:    "heatbeat",
				Content: "pong",
			}
			pongByte, err := proto.Marshal(pong)
			if err != nil {
				log.Println("client write message error")
				c.Ws.Close()
				break
			}
			c.Ws.WriteMessage(websocket.BinaryMessage, pongByte)
		} else {
			
			// 将消息送入通道
			ServerInstance.Broadcast <- message
			// 将消息存入数据库
			dao.DB.Table(model.Message{}.GetTable()).Create(&model.Message{
				SenderID:  pbMessage.From,
				ReceiveID: pbMessage.To,
				Content:   pbMessage.Content,
				Status:    0,
			})
		}
	}
}

func (c *Client) Write() {
	defer func() {
		c.Ws.Close()
	}()
	for message := range c.Send {
		c.Ws.WriteMessage(websocket.BinaryMessage, message)
	}
}
