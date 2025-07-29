package test

import (
	"cchat/internal/dao"
	"cchat/internal/dto"
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func TestRedis(t *testing.T) {
	dao.InitRedis()
	dao.REDIS.Set(context.Background(), "test", "123", 0)
	val, err := dao.REDIS.Get(context.Background(), "test").Result()
	if err != nil {
		t.Errorf("get redis failed, err: %v", err)
	}
	fmt.Println(val)
}

func TestInsertFriend(t *testing.T) {
	dao.InitRedis()
	pipe := dao.REDIS.Pipeline()
	ctx := context.Background()
	friend := []dto.FriendInfo{}
	friend = append(friend, dto.FriendInfo{
		FriendUUID:     "7dc1adea-72f4-3c27-a887-fd766c4eb44b",
		FriendName:     "wanger",
		FriendAvatar:   "d2971ab2-ae4b-4192-96cd-639edd582002.png",
		FriendNickname: "cccc",
		Status:         1,
	})
	friend = append(friend, dto.FriendInfo{
		FriendUUID:     "4dfc6b14-7213-3363-8009-b23c56e3a1b1",
		FriendName:     "123",
		FriendAvatar:   "d2971ab2-ae4b-4192-96cd-639edd582002.png",
		FriendNickname: "不忘可乐",
		Status:         1,
	})
	// 每个好友存储为哈希字段
	for _, f := range friend {
		// 序列化结构体为JSON
		friendData, _ := json.Marshal(f)

		// 使用用户ID作为键，好友UUID作为字段
		pipe.HSet(ctx,
			"0704e879-c1e8-37f8-8a71-79c7e122f474",
			f.FriendUUID,
			string(friendData),
		)
	}

	// 批量执行
	pipe.Exec(ctx)
}

func TestGetFriendInfo(t *testing.T) {
	dao.InitRedis()
	result, err := dao.REDIS.HGetAll(
		context.Background(),
		"0704e879-c1e8-37f8-8a71-79c7e122f474",
	).Result()
	if err != nil {
		t.Errorf("get redis failed, err: %v", err)
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

func TestClearRedis(t *testing.T) {
	dao.InitRedis()
	dao.REDIS.Del(context.Background(), "0704e879-c1e8-37f8-8a71-79c7e122f474")
}
func TestShowAllKeys(t *testing.T) {
	dao.InitRedis()
	keys, err := dao.REDIS.Keys(context.Background(), "*").Result()
	if err != nil {
		t.Errorf("get redis failed, err: %v", err)
	}
	fmt.Println(keys)
}
