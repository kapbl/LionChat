package service

import (
	"cchat/pkg/logger"
	"cchat/pkg/protocol"
	"log"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Client struct {
	Ws         *websocket.Conn
	UUID       string
	Send       chan []byte
	mutex      sync.RWMutex
	done       chan struct{}
	ConnTime   int64
	RemoteAddr string
}

func NewClient(name string, ws *websocket.Conn) *Client {
	return &Client{
		Ws:         ws,
		UUID:       name,
		Send:       make(chan []byte),
		mutex:      sync.RWMutex{},
		done:       make(chan struct{}),
		ConnTime:   time.Now().Unix(),
		RemoteAddr: ws.RemoteAddr().String(),
	}
}

func (c *Client) Read() {
	defer func() {
		//c.mutex.Lock()
		close(c.done)
		ServerInstance.Ungister <- c
		c.Ws.Close()
		//c.mutex.Unlock()
	}()

	for {
		select {
		case <-c.done:
			return
		default:
			//c.mutex.RLock()
			_, message, err := c.Ws.ReadMessage()
			//c.mutex.RUnlock()

			if err != nil {
				log.Println("client read message error:", err)
				return
			}

			// 将消息编码成protobuf
			pbMessage := protocol.Message{}
			if err := proto.Unmarshal(message, &pbMessage); err != nil {
				logger.Error("failed to unmarshal message:", zap.Error(err))
				continue
			}

			// 心跳消息
			if pbMessage.Type == "heatbeat" {
				pong := &protocol.Message{
					Type:    "heatbeat",
					Content: "pong",
				}
				pongByte, err := proto.Marshal(pong)
				if err != nil {
					log.Println("failed to marshal pong message:", err)
					continue
				}

				select {
				case c.Send <- pongByte:
					// 心跳响应发送成功
				case <-c.done:
					return
				default:
					log.Println("failed to send pong message: channel full")
					return
				}
			} else {
				// 将消息送入通道
				select {
				case ServerInstance.Broadcast <- message:
					// 消息已送入广播通道
				case <-c.done:
					return
				default:
					logger.Error("failed to broadcast message: channel full")
					return
				}
			}
		}
	}
}

func (c *Client) Write() {
	defer func() {
		//c.mutex.Lock()
		c.Ws.Close()
		//c.mutex.Unlock()
	}()

	for {
		select {
		case <-c.done:
			return
		case message, ok := <-c.Send:
			if !ok {
				return
			}
			//c.mutex.Lock()
			err := c.Ws.WriteMessage(websocket.BinaryMessage, message)
			//c.mutex.Unlock()

			if err != nil {
				log.Println("failed to write message:", err)
				return
			}
		}
	}
}
