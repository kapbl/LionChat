package model

import "gorm.io/gorm"

type Message struct {
	gorm.Model        // 嵌入标准字段（ID/CreatedAt/UpdatedAt/DeletedAt）
	SenderID   string `gorm:"index:idx_sender;not null;comment:'发送者ID'"`
	ReceiveID  string `gorm:"index:idx_receiver;not null;comment:'接收者ID'"`
	Content    string `gorm:"type:text;not null;comment:'消息内容'"`
	Status     int8   `gorm:"type:tinyint(1);default:0;comment:'0未读 1已读 2撤回'"`
	MessageID  string `gorm:"type:varchar(255);primaryKey;comment:'消息唯一标识符'"`
}

func (Message) GetTable() string {
	return "message"
}
