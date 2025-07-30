package dao

import (
	"cchat/pkg/logger"
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

var REDIS *redis.Client

// InitRedis 初始化Redis连接
func InitRedis(addr, password string, db, poolSize, minIdleConns int) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		// 连接池超时配置
		DialTimeout:  5 * time.Second,  // 连接超时
		ReadTimeout:  3 * time.Second,  // 读取超时
		WriteTimeout: 3 * time.Second,  // 写入超时
		PoolTimeout:  4 * time.Second,  // 池超时
		IdleTimeout:  5 * time.Minute,  // 空闲连接超时
	})

	REDIS = rdb

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := REDIS.Ping(ctx).Err(); err != nil {
		logger.Fatal("连接Redis失败",
			zap.Error(err),
			zap.String("addr", addr),
			zap.Int("db", db))
	}

	logger.Info("Redis连接成功",
		zap.String("addr", addr),
		zap.Int("db", db),
		zap.Int("pool_size", poolSize),
		zap.Int("min_idle_conns", minIdleConns))
}

// DeleteCache 删除缓存
func DeleteCache(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := REDIS.Del(ctx, key).Err()
	if err != nil {
		logger.Error("删除Redis缓存失败",
			zap.Error(err),
			zap.String("key", key))
	}
}

type Action func() error

// UpdataUserWithDelayDoubleDelete 使用延迟双删策略更新用户数据
func UpdataUserWithDelayDoubleDelete(user_uuid string, action Action) error {
	// 第一次删除缓存
	DeleteCache(user_uuid)

	// 更新数据库
	if err := action(); err != nil {
		return err
	}

	// 延迟删除，防止缓存双写不一致
	time.Sleep(500 * time.Millisecond)

	// 第二次删除缓存
	DeleteCache(user_uuid)
	return nil
}
