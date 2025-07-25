package dao

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

func InitDB(dsn string) error {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: &schema.NamingStrategy{
			NoLowerCase:   false,
			SingularTable: true,
		},
	})
	if err != nil {
		log.Fatalf("init db failed, err: %v", err)
		return err
	}
	DB = db
	return nil
}
