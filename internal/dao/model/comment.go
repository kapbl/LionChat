package model

import "time"

type Comment struct {
	ID         int64     `gorm:"primaryKey"`
	MomentID   int64     `gorm:"column:moment_id"`
	UserID     int64     `gorm:"column:user_id"`
	Content    string    `gorm:"column:content"`
	CreateTime time.Time `gorm:"column:create_time"`
}

func (Comment) TableName() string {
	return "comment"
}
