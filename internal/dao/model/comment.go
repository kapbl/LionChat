package model

import "time"

type Comment struct {
	ID        int64      `gorm:"primaryKey"`
	MomentID  int64      `gorm:"column:moment_id"`
	UserID    int64      `gorm:"column:user_id"`
	Content   string     `gorm:"column:content"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at;default:null"`
}

func (Comment) TableName() string {
	return "comment"
}
