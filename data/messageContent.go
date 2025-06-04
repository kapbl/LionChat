package jsoncontent

type MessageContent struct{
	MessageType int `json:"message_type"`// 0：单聊 1：群聊
	Content string `json:"content"`// 消息的内容
	From string `json:"from"`// 发送方: uuid
	To string `json:"to"`// 接收方: uuid
}