package model

import "time"

type UserFriends struct {
	Id       int        `gorm:"column:id;primaryKey"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	UserID   int        `gorm:"column:user_id"`
	FriendID int        `gorm:"column:friend_id"`
	Status   int        `gorm:"column:status"`
}

func (UserFriends) GetTable() string {
	return "user_friends"
}
