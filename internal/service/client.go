package service

import (
	"cchat/pkg/logger"
	"cchat/pkg/protocol"
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
	LoginTime  int64
	Heartbeat  int64
	RemoteAddr string
}

func NewClient(ws *websocket.Conn, uuid string) *Client {
	return &Client{
		Ws:         ws,
		UUID:       uuid,
		Send:       make(chan []byte),
		mutex:      sync.RWMutex{},
		done:       make(chan struct{}),
		ConnTime:   time.Now().Unix(),
		LoginTime:  time.Now().Unix(),
		Heartbeat:  time.Now().Unix(),
		RemoteAddr: ws.RemoteAddr().String(),
	}
}

func (c *Client) Read() {
	defer func() {
		close(c.done)
		ServerInstance.Ungister <- c
		c.Ws.Close()
	}()

	// 设置读取超时和心跳处理
	c.Ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Ws.SetPongHandler(func(string) error {
		c.Ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.Heartbeat = time.Now().Unix()
		logger.Debug("received pong from client", zap.String("uuid", c.UUID))
		return nil
	})

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
		// 将消息编码成protobuf
		pbMessage := protocol.Message{}
		if err := proto.Unmarshal(message, &pbMessage); err != nil {
			logger.Error("failed to unmarshal message:", zap.Error(err))
			continue
		}
		logger.Debug("client read message:", zap.String("message", pbMessage.Content))

		if ServerInstance.Broadcast == nil {
			logger.Error("server instance is nil")
			return
		}
		// 将消息送入通道
		select {
		case ServerInstance.Broadcast <- message:
			logger.Debug("client send message to server")
		case <-c.done:
			return
		case <-time.After(5 * time.Second):
			logger.Error("broadcast message timeout")
			return
		default:
			logger.Error("failed to broadcast message: channel full")
			return
		}
	}
}

func (c *Client) Write() {
	defer func() {
		c.Ws.Close()
	}()

	// 创建心跳定时器，每30秒发送一次ping
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-c.done:
			return
		case message, ok := <-c.Send:
			if !ok {
				return
			}
			c.Ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := c.Ws.WriteMessage(websocket.BinaryMessage, message)
			if err != nil {
				logger.Error("failed to write message:", zap.Error(err))
				return
			}
		case <-pingTicker.C:
			// 发送ping消息保持连接活跃
			c.Ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				logger.Error("failed to send ping:", zap.Error(err))
				return
			}
			logger.Debug("sent ping to client", zap.String("uuid", c.UUID))
		}
	}
}
