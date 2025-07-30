package dao

import (
	"cchat/pkg/logger"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(dsn string, maxIdleConns, maxOpenConns int, connMaxLifetime time.Duration) error {
	config := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
	}

	db, err := gorm.Open(mysql.Open(dsn), config)
	if err != nil {
		logger.Fatal("连接数据库失败",
			zap.Error(err),
			zap.String("dsn", dsn))
		return err
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("获取数据库实例失败", zap.Error(err))
		return err
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(maxIdleConns)         // 设置空闲连接池中的最大连接数
	sqlDB.SetMaxOpenConns(maxOpenConns)         // 设置打开数据库连接的最大数量
	sqlDB.SetConnMaxLifetime(connMaxLifetime)   // 设置连接可复用的最大时间

	DB = db
	logger.Info("数据库连接池配置成功",
		zap.Int("max_idle_conns", maxIdleConns),
		zap.Int("max_open_conns", maxOpenConns),
		zap.Duration("conn_max_lifetime", connMaxLifetime))

	return nil
}
