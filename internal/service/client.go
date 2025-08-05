package service

import (
	"cchat/pkg/logger"
	"cchat/pkg/protocol"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Client struct {
	Ws              *websocket.Conn // 客户端连接
	UUID            string          // 客户端唯一标识
	Send            chan []byte     // 客户端发送消息通道
	done            chan struct{}
	ConnTime        int64
	LoginTime       int64
	Heartbeat       int64
	RemoteAddr      string
	FragmentManager *FragmentManager // 客户端分片管理器
}

func NewClient(ws *websocket.Conn, uuid string) *Client {
	return &Client{
		Ws:              ws,
		UUID:            uuid,
		Send:            make(chan []byte),
		done:            make(chan struct{}),
		ConnTime:        time.Now().Unix(),
		LoginTime:       time.Now().Unix(),
		Heartbeat:       time.Now().Unix(),
		RemoteAddr:      ws.RemoteAddr().String(),
		FragmentManager: NewFragmentManager(),
	}
}

func (c *Client) Read() {
	defer func() {
		close(c.done)
		ServerInstance.Ungister <- c
		c.Ws.Close()
	}()
	for {
		_, message, err := c.Ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("client read message error:", zap.Error(err))
			} else {
				logger.Info("client disconnected normally", zap.String("uuid", c.UUID), zap.Error(err))
			}
			return
		}
		msg := &protocol.Message{}
		proto.Unmarshal(message, msg)
		if msg.Type == "ping" {
			pong := &protocol.Message{
				Content: "pong",
				Type:    "ping",
			}
			pongByte, err2 := proto.Marshal(pong)
			if nil != err2 {
				logger.Error("client marshal message error", zap.Error(err2))
			}
			c.Ws.WriteMessage(websocket.BinaryMessage, pongByte)
		} else {
			if ServerInstance.Broadcast == nil {
				logger.Error("server instance is nil")
				return
			}
			ServerInstance.Broadcast <- message
		}
	}
}

func (c *Client) Write() {
	defer func() {
		c.Ws.Close()
	}()
	for message := range c.Send {
		err := c.Ws.WriteMessage(websocket.BinaryMessage, message)
		if err != nil {
			logger.Error("failed to write message:", zap.Error(err))
			return
		}
	}
}
