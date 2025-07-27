package model

import "time"

type Group struct {
	Id          int       `gorm:"primary_key;AUTO_INCREMENT;comment:'id'"`
	Name        string    `gorm:"type:varchar(150);not null;unique_index:idx_name;comment:'组名'"`
	Desc        string    `gorm:"type:varchar(255);comment:'描述'"`
	MemberCount int       `gorm:"default:0;comment:'成员数量'"`
	OwnerId     int       `gorm:"comment:'群主id'"`
	CreateAt    time.Time `gorm:"comment:'创建时间'"`
	UpdateAt    time.Time `gorm:"comment:'更新时间'"`
	DeleteAt    time.Time `gorm:"comment:'删除时间'"`
}

type GroupMember struct {
	Id       int       `gorm:"primary_key;AUTO_INCREMENT;comment:'id'"`
	GroupId  int       `gorm:"comment:'组id'"`
	UserId   int       `gorm:"comment:'用户id'"`
	CreateAt time.Time `gorm:"comment:'创建时间'"`
	UpdateAt time.Time `gorm:"comment:'更新时间'"`
	DeleteAt time.Time `gorm:"comment:'删除时间';default:null"`
}
