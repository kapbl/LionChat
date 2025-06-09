package main

import (
	"chatLion/database"
	"chatLion/model"
	"log"

	"gorm.io/gorm"
)

func main() {
	dsn := "root:0220059cyCY@tcp(127.0.0.1:3306)/chatLion?charset=utf8mb4&parseTime=True&loc=Local"
	db := database.InitGorm(dsn)
	db.AutoMigrate(&model.UserModel{}, &model.Group{}, &model.UserFriend{})

	//testQueryMyFriends(db)
}

func testQueryMyFriends(db *gorm.DB) {
	var friends []model.UserFriend
	// 查询user_id = 5的所有好友
	db.Where("user_id = ?", 5).Find(&friends)
	log.Println(friends)
}

func newFunction(db *gorm.DB) model.Group {
	var group model.Group
	var user model.UserModel
	db.First(&group, "name = ?", "game")          // 查询记录
	db.First(&user, "email = ?", "libai@163.com") // 查询记录
	err := db.Model(&group).Association("Members").Delete(&user)
	if err != nil {
		log.Println(err)
	}
	return group
}

func testAddGroup(db *gorm.DB) {
	group := model.Group{
		Name: "Lions",
		Members: []model.UserModel{
			{Email: " libai@163.com "}, // 仅需提供用户ID
			{Email: " lishangyin@163.com"},
			{Email: " dufu@163.com"},
		},
	}
	db.Create(&group) // 自动写入 groups 和 group_members

}

func testQueryGroup(db *gorm.DB) {
	// 查询并打印组信息
	var fetchedGroup model.Group
	// 根据组名查询一个组
	db.Preload("Members").First(&fetchedGroup, "name = ?", "篮球1") // 预加载成员信息
	log.Printf("Group: %+v", fetchedGroup)
	// db.Preload("Members").First(&fetchedGroup, 1) // 预加载成员信息
	// log.Printf("Group: %+v", fetchedGroup)
	// // 打印组的成员信息
	// for _, member := range fetchedGroup.Members {
	// 	log.Printf("Member: %+v", member)
	// }
}

func testDeleteGroup(db *gorm.DB) {
	// db.Model(&group).Association("Members").Delete(&model.UserModel{Email: "user001"})

}
