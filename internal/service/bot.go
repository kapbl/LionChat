package service

type BotClient struct {
	// Ws       *websocket.Conn // 客户端连接
	Username string
	UUID     string      // 客户端唯一标识
	Send     chan []byte // 客户端发送消息通道，
	WorkerID int         // 服务自己的worker id
}

func NewBotClient(username string, uuid string, workerID int) *BotClient {

	// 为机器人设置一个websocket连接
	// ws, _, err := websocket.DefaultDialer.Dial("ws://localhost:9922", nil)
	// if err != nil {
	// 	logger.Error("bot client dial error:", zap.Error(err))
	// 	return nil
	// }
	bot := &BotClient{
		Username: username,
		// Ws:       nil,
		UUID:     uuid,
		Send:     make(chan []byte, 100),
		WorkerID: workerID,
	}
	return bot
}

// // 客户端读取一个消息
// func (c *BotClient) Read() {
// 	for {
// 		_, message, err := c.Ws.ReadMessage()
// 		if err != nil {
// 			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
// 				logger.Error("client read message error:", zap.Error(err))
// 			} else {
// 				logger.Info("client disconnected normally", zap.String("uuid", c.UUID), zap.Error(err))
// 			}
// 			return
// 		}
// 		msg := &protocol.Message{}
// 		proto.Unmarshal(message, msg)
// 		if msg.Type == "ping" {
// 			pong := &protocol.Message{
// 				Content: "pong",
// 				Type:    "ping",
// 			}
// 			pongByte, err2 := proto.Marshal(pong)
// 			if nil != err2 {
// 				logger.Error("client marshal message error", zap.Error(err2))
// 			}
// 			c.Ws.WriteMessage(websocket.BinaryMessage, pongByte)
// 		} else {
// 			// 应该向服务自己的工作者告诉我读到了一个消息，请帮我转发， 具体转发到 to 的client
// 			worker := ServerInstance.WorkerHouse.GetWorkerByID(c.WorkerID)
// 			if worker == nil {
// 				logger.Error("worker not found")
// 				return
// 			}
// 			worker.Broadcast <- message
// 		}
// 	}
// }

// // 客户端写入消息
// func (c *BotClient) Write() {
// 	for message := range c.Send {
// 		err := c.Ws.WriteMessage(websocket.BinaryMessage, message)
// 		if err != nil {
// 			logger.Error("failed to write message:", zap.Error(err))
// 			return
// 		}
// 	}
// }
