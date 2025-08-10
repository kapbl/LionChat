package model

import "time"

type Comment struct {
	ID       int64     `gorm:"primaryKey"`
	MomentID int64     `gorm:"column:moment_id"`
	UserID   int64     `gorm:"column:user_id"`
	Content  string    `gorm:"column:content"`
	CreateAt time.Time `gorm:"column:create_at"`
}

func (Comment) TableName() string {
	return "comment"
}
