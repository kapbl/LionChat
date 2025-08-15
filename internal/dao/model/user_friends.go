package model

import "time"

type UserFriends struct {
	Id       int        `gorm:"column:id;primaryKey"`
	CreateAt time.Time  `gorm:"column:create_at"`
	UpdateAt time.Time  `gorm:"column:update_at"`
	DeleteAt *time.Time `gorm:"column:delete_at"`
	UserID   int        `gorm:"column:user_id"`
	FriendID int        `gorm:"column:friend_id"`
	Status   int        `gorm:"column:status"`
}

func (UserFriends) GetTable() string {
	return "user_friends"
}
