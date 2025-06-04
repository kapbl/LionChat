package model

import "gorm.io/gorm"

type UserModel struct {
	gorm.Model
	// 用户昵称可以表示
	Nickname string `gorm:"size:32"`
	// 用户邮箱是唯一的
	Email string `gorm:"size:64;uniqueIndex"`
	// 用户密码
	Password string      `gorm:"size:128"`
	Friends  []UserModel `gorm:"many2many:user_friends;"` // 多对多关联
}
