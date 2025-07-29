package dao

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

var REDIS *redis.Client

func InitRedis() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	REDIS = rdb
	if err := REDIS.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("连接redis失败, err: %v", err)
	}
}

// DeleteCache 删除缓存
func DeleteCache(key string) {
	err := REDIS.Del(context.Background(), key).Err()
	if err != nil {
		fmt.Println("删除 Redis 缓存失败：", err)
	}
}

type Action func() error

func UpdataUserWithDelayDoubleDelete(user_uuid string, action Action) error {
	DeleteCache(user_uuid)
	// 更新数据库
	if err := action(); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)

	// 第二次删除缓存（防止脏数据回写）
	DeleteCache(user_uuid)
	return nil
}
