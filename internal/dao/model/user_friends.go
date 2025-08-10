package model

import "time"

type UserFriends struct {
	CreateAt time.Time  `json:"create_at"`
	UpdateAt time.Time  `json:"update_at"`
	DeleteAt *time.Time `json:"delete_at"`
	UserID   int        `gorm:"comment:'用户id'"`
	FriendID int        `gorm:"comment:'好友id'"`
	Status   int        `gorm:"comment:'状态：1或0'"`
}

func (UserFriends) GetTable() string {
	return "user_friends"
}
