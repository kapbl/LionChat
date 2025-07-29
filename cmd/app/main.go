package main

import (
	"cchat/config"
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/internal/router"
	"cchat/internal/service"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"
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
	dao.InitRedis()
	// 自动迁移表
	dao.DB.AutoMigrate(&model.Users{}, &model.UserFriends{},
		&model.Message{}, &model.Group{}, &model.GroupMember{})
	// 初始化路由
	router.InitWebEngine()
	// 启动服务器
	go service.ServerInstance.Start()
	//go monitorGoroutines()
	// 优雅关闭处理
	setupGracefulShutdown()
	// 启动路由
	router.RunEngine()
}

func monitorGoroutines() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			f, _ := os.Create("goroutine.prof")
			pprof.Lookup("goroutine").WriteTo(f, 1)
			f.Close()
		}
	}
}

// 新增优雅关闭处理
func setupGracefulShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("接收到关闭信号，正在清理资源...")
		// 这里可以添加数据库关闭等清理操作
		os.Exit(0)
	}()
}
