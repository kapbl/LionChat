package model

// 中间表模型
type UserFriend struct {
	UserID   uint `gorm:"primaryKey"`
	FriendID uint `gorm:"primaryKey"`
}
