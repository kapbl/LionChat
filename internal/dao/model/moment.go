package model

import "time"

type Moment struct {
	ID        int64      `gorm:"primaryKey"` // 该记录的主键
	UserID    int64      `gorm:"column:user_id"`
	Content   string     `gorm:"column:content"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

func (Moment) TableName() string {
	return "moment"
}

type Timeline struct {
	ID        int64      `gorm:"primaryKey"`
	UserID    int64      `gorm:"column:user_id"`
	MomentID  int64      `gorm:"column:moment_id"`
	IsOwn     bool       `gorm:"column:is_own"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
}

func (Timeline) TableName() string {
	return "timeline"
}
