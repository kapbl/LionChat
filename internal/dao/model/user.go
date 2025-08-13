package model

import "time"

type Users struct {
	Id            int32      `gorm:"primary_key;AUTO_INCREMENT;comment:'id'"`
	Uuid          string     `gorm:"type:varchar(150);not null;unique_index:idx_uuid;comment:'uuid'"`
	Username      string     `gorm:"unique;not null; comment:'用户名'"`
	Password      string     `gorm:"type:varchar(150);not null; comment:'密码'"`
	Nickname      string     `gorm:"comment:'昵称'"`
	Avatar        string     `gorm:"type:varchar(150);comment:'头像'"`
	Email         string     `gorm:"type:varchar(80);column:email;comment:'邮箱'"`
	GroupVersion  int        `gorm:"comment:'控制用户的群组版本号，用于防止Redis读取到过期的数据'"`
	FriendVersion int        `gorm:"comment:'控制用户的好友版本号，用于防止Redis读取到过期的数据'"`
	CreateAt      time.Time  `gorm:"created_at"`
	UpdateAt      time.Time  `gorm:"updated_at"`
	DeleteAt      *time.Time `gorm:"deleted_at"`
}

func (Users) GetTable() string {
	return "users"
}
