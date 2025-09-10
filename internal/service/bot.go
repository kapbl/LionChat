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
	bot := &BotClient{
		Username: username,
		// Ws:       nil,
		UUID:     uuid,
		Send:     make(chan []byte, 100),
		WorkerID: workerID,
	}
	return bot
}
