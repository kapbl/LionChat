package main

import (
	"cchat/config"
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/internal/router"
	"cchat/internal/service"
	"fmt"
	"log"
)

func main() {
	fmt.Printf("%s", "服务启动")
	// 加载配置
	appConfig := config.LoadConfig()
	// 初始化数据库
	if err := dao.InitDB(appConfig.MySQL.DSN); err != nil {
		log.Fatalf("数据库初始化失败，错误信息：%v", err)
	} else {
		log.Printf("数据库初始化成功")
	}
	// 自动迁移表
	dao.DB.AutoMigrate(&model.User{}, &model.UserFriends{},
		&model.Message{}, &model.Group{}, &model.GroupMember{})
	// 初始化路由
	router.InitWebEngine()
	// 启动服务器
	go service.ServerInstance.Start()
	// 启动路由
	router.RunEngine()
}
