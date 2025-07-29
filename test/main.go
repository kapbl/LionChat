package main

import (
	"cchat/internal/dao"
	"cchat/internal/dto"
	"context"
	"encoding/json"
	"fmt"
)

func main() {
	// TestClearRedis()
	TestGetFriendInfo()
}
func TestClearRedis() {
	dao.InitRedis()
	dao.REDIS.Del(context.Background(), "0704e879-c1e8-37f8-8a71-79c7e122f474")
}

func TestGetFriendInfo() {
	dao.InitRedis()
	result, err := dao.REDIS.HGetAll(
		context.Background(),
		"0704e879-c1e8-37f8-8a71-79c7e122f474",
	).Result()
	if err != nil {
		fmt.Errorf("get redis failed, err: %v", err)
	}
	friends := make([]dto.FriendInfo, 0)
	for _, v := range result {
		var friend dto.FriendInfo
		if err := json.Unmarshal([]byte(v), &friend); err != nil {
			continue // 或返回错误
		}
		friends = append(friends, friend)
	}
	fmt.Println(friends)
}
