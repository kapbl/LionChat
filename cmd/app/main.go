package main

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/internal/router"
	"cchat/internal/service"
	"cchat/pkg/config"
	"cchat/pkg/logger"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	// 初始化日志
	logger.InitLogger()
	// 加载配置
	appConfig := config.LoadConfig()
	// 初始化数据库
	if err := dao.InitDB(appConfig.MySQL.DSN, appConfig.MySQL.MaxIdleConns, appConfig.MySQL.MaxOpenConns, time.Duration(appConfig.MySQL.ConnMaxLifetime)); err != nil {
		logger.Fatal("数据库初始化失败", zap.Error(err))
	} else {
		logger.Info("数据库初始化成功")
	}
	// 初始化Redis
	if err := dao.InitRedis(appConfig.Redis.Addr, appConfig.Redis.Password, appConfig.Redis.DB, appConfig.Redis.PoolSize, appConfig.Redis.MinIdleConns); err != nil {
		logger.Fatal("Redis初始化失败", zap.Error(err))
	} else {
		logger.Info("Redis初始化成功")
	}
	// 初始化Kafka
	if err := dao.InitKafka(&appConfig); err != nil {
		logger.Error("Kafka初始化失败", zap.Error(err))
		// Kafka初始化失败不影响主服务启动，只记录错误
	} else {
		logger.Info("Kafka初始化成功")
	}
	// 自动迁移表
	dao.DB.AutoMigrate(&model.Users{}, &model.UserFriends{},
		&model.Message{}, &model.Group{}, &model.GroupMember{})
	// 初始化路由
	router.InitWebEngine()
	// 启动Kafka消费者服务
	if dao.KafkaConsumerInstance != nil {
		kafkaConsumerService, err := service.NewKafkaConsumerService(&appConfig)
		if err != nil {
			logger.Error("创建Kafka消费者服务失败", zap.Error(err))
		} else {
			go kafkaConsumerService.Start()
			logger.Info("Kafka消费者服务启动成功")
		}
	}
	// 启动服务器
	go service.ServerInstance.Start()
	// 优雅关闭处理
	setupGracefulShutdown()
	// 启动路由
	router.RunEngine(&appConfig)
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

// setupGracefulShutdown 优雅关闭处理
func setupGracefulShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("接收到关闭信号，正在清理资源...")

		// 关闭Kafka连接
		if dao.KafkaProducerInstance != nil {
			if err := dao.KafkaProducerInstance.Close(); err != nil {
				logger.Error("关闭Kafka生产者失败", zap.Error(err))
			} else {
				logger.Info("Kafka生产者已关闭")
			}
		}

		if dao.KafkaConsumerInstance != nil {
			if err := dao.KafkaConsumerInstance.Close(); err != nil {
				logger.Error("关闭Kafka消费者失败", zap.Error(err))
			} else {
				logger.Info("Kafka消费者已关闭")
			}
		}

		// 关闭数据库连接
		if sqlDB, err := dao.DB.DB(); err == nil {
			sqlDB.Close()
		}

		// 关闭Redis连接
		if dao.REDIS != nil {
			dao.REDIS.Close()
		}

		// 等待一段时间确保资源清理完成
		time.Sleep(time.Second * 2)
		os.Exit(0)
	}()
}
