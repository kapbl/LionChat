package model

import "time"

type Moment struct {
	ID         int64      `gorm:"primaryKey"` // 该记录的主键
	UserID     int64      `gorm:"column:user_id"`
	Content    string     `gorm:"column:content"`
	CreateTime time.Time  `gorm:"column:create_time"`
	DeleteTime *time.Time `gorm:"column:delete_time"`
}

func (Moment) TableName() string {
	return "moment"
}

type Timeline struct {
	ID         int64     `gorm:"primaryKey"`
	UserID     int64     `gorm:"column:user_id"`
	MomentID   int64     `gorm:"column:moment_id"`
	IsOwn      bool      `gorm:"column:is_own"`
	CreateTime time.Time `gorm:"column:create_time"`
}

func (Timeline) TableName() string {
	return "timeline"
}
