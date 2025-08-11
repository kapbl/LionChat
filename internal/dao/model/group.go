package model

import "time"

type Group struct {
	Id          int        `gorm:"primary_key;AUTO_INCREMENT;comment:'id'"`
	UUID        string     `gorm:"type:varchar(150);not null;unique_index:idx_name;comment:'UUID'"`
	Name        string     `gorm:"type:varchar(150);not null;unique_index:idx_name;comment:'组名'"`
	Desc        string     `gorm:"type:varchar(255);comment:'描述'"`
	Type        string     `gorm:"type:varchar(255);comment:'类型'"`
	MemberCount int        `gorm:"default:0;comment:'成员数量'"`
	OwnerId     int        `gorm:"comment:'群主id'"`
	CreateAt    time.Time  `gorm:"comment:'创建时间'"`
	UpdateAt    time.Time  `gorm:"comment:'更新时间'"`
	DeleteAt    *time.Time `gorm:"comment:'删除时间'"`
}

type GroupMember struct {
	Id        int        `gorm:"primary_key;AUTO_INCREMENT;comment:'id'"`
	GroupId   int        `gorm:"column:group_id;comment:'组id'"`
	GroupUUID string     `gorm:"column:group_uuid;comment:'组UUID'"`
	UserId    int        `gorm:"column:user_id;comment:'用户id'"`
	UserUUID  string     `gorm:"column:user_uuid;comment:'用户UUID'"`
	CreateAt  time.Time  `gorm:"column:create_at;comment:'创建时间'"`
	UpdateAt  time.Time  `gorm:"column:update_at;comment:'更新时间'"`
	DeleteAt  *time.Time `gorm:"column:delete_at;comment:'删除时间';default:null"`
}
