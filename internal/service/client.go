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
	Ws         *websocket.Conn // 客户端连接
	UUID       string          // 客户端唯一标识
	Send       chan []byte     // 客户端发送消息通道，
	done       chan struct{}   // 是否表示客户端已断开
	ConnTime   int64           // 连接那一时间戳
	RemoteAddr string          // 远程地址
	workerID   int             // 服务自己的worker id
}

// 新建一个客户端
func NewClient(ws *websocket.Conn, uuid string) *Client {
	return &Client{
		Ws:         ws,
		UUID:       uuid,
		Send:       make(chan []byte),
		done:       make(chan struct{}),
		ConnTime:   time.Now().Unix(),
		RemoteAddr: ws.RemoteAddr().String(),
	}
}

// 客户端读取一个消息
func (c *Client) Read() {
	defer func() {
		close(c.done)
		worker := ServerInstance.WorkerHouse.GetWorkerByID(c.workerID)
		if worker == nil {
			logger.Error("worker not found")
			return
		}
		worker.Unregister <- c
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
			// 应该向服务自己的工作者告诉我读到了一个消息，请帮我转发
			worker := ServerInstance.WorkerHouse.GetWorkerByID(c.workerID)
			if worker == nil {
				logger.Error("worker not found")
				return
			}
			worker.Broadcast <- message
		}
	}
}

// 客户端写入消息
func (c *Client) Write() {
	defer func() {
		c.Ws.Close()
		worker := ServerInstance.WorkerHouse.GetWorkerByID(c.workerID)
		if worker == nil {
			logger.Error("worker not found")
			return
		}
		worker.Unregister <- c

	}()
	for message := range c.Send {
		err := c.Ws.WriteMessage(websocket.BinaryMessage, message)
		if err != nil {
			logger.Error("failed to write message:", zap.Error(err))
			return
		}
	}
}
