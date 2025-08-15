package model

import (
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) {
	db.AutoMigrate(&Users{}, &UserFriends{}, &Group{}, &GroupMember{}, &Moment{}, &Timeline{}, &Like{}, &Comment{})
}
