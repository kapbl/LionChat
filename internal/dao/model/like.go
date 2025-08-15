package model

import "time"

type Like struct {
	MomentID  int64      `gorm:"column:moment_id"`
	UserID    int64      `gorm:"column:user_id"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

func (l *Like) TableName() string {
	return "like"
}
