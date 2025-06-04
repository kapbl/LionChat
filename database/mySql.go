package database

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitGorm(dataSouce string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dataSouce), &gorm.Config{})
	if err != nil {
		log.Println(err)
		return nil
	}
	log.Println("数据库连接成功")
	return db
}

// 初始化数据库表
// func main() {
// 	// 初始化数据库
// 	dsn := "root:0220059cyCY@tcp(127.0.0.1:3306)/chatLion?charset=utf8mb4&parseTime=True&loc=Local"
// 	db := InitGorm(dsn)
// 	err := db.AutoMigrate(&model.UserModel{})
// 	if err != nil {
// 		log.Println("表结构生成失败")
// 	}
// 	log.Println("表结构生成成功")
// }
