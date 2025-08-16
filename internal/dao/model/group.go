package model

import "time"

type Group struct {
	Id          int        `gorm:"column:id;primary_key;AUTO_INCREMENT;comment:'id'"`
	UUID        string     `gorm:"column:uuid;type:varchar(150);not null;unique_index:idx_name;comment:'UUID'"`
	Name        string     `gorm:"column:name;type:varchar(150);not null;unique_index:idx_name;comment:'组名'"`
	Desc        string     `gorm:"column:desc;type:varchar(255);comment:'描述'"`
	Type        string     `gorm:"column:type;type:varchar(255);comment:'类型'"`
	MemberCount int        `gorm:"column:member_count;default:0;comment:'成员数量'"`
	OwnerId     int        `gorm:"column:owner_id;comment:'群主id'"`
	CreatedAt   time.Time  `gorm:"column:created_at;comment:'创建时间'"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;comment:'更新时间'"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;comment:'删除时间';default:null"`
}

type GroupMember struct {
	Id        int        `gorm:"column:id;primary_key;AUTO_INCREMENT;comment:'id'"`
	GroupId   int        `gorm:"column:group_id;comment:'组id'"`
	GroupUUID string     `gorm:"column:group_uuid;comment:'组UUID'"`
	UserId    int        `gorm:"column:user_id;comment:'用户id'"`
	UserUUID  string     `gorm:"column:user_uuid;comment:'用户UUID'"`
	CreatedAt time.Time  `gorm:"column:created_at;comment:'创建时间'"`
	UpdatedAt time.Time  `gorm:"column:updated_at;comment:'更新时间'"`
	DeletedAt *time.Time `gorm:"column:deleted_at;comment:'删除时间';default:null"`
}
