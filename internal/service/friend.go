package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/internal/dto"
	"cchat/pkg/cerror"
	"cchat/pkg/logger"
	"cchat/pkg/protocol"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
	"gorm.io/gorm"
)

type FriendService struct {
	UserID   int64
	UUID     string
	Username string
	db       *gorm.DB
}

func NewFriendService(userID int64, uuid string, username string, db *gorm.DB) *FriendService {
	return &FriendService{
		UserID:   userID,
		UUID:     uuid,
		db:       db,
		Username: username,
	}
}

// ✅
func (f *FriendService) SearchClient(information string) (*dto.UserInfo, *cerror.CodeError) {
	user := model.Users{}
	err := dao.DB.Table(user.GetTable()).Where("username = ? OR nickname = ? OR email = ?", information, information, information).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cerror.NewCodeError(4444, "用户不存在")
		}
		return nil, cerror.NewCodeError(4444, err.Error())
	}
	resp := &dto.UserInfo{
		Username: user.Username,
		Email:    user.Email,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
	}
	return resp, nil
}

// ✅
func (f *FriendService) AddFriend(req *dto.AddFriendRequest) *cerror.CodeError {
	// 离线用户， 不经过websocket
	// 在线用户， 经过websocket
	// 根据信息查询uuid
	// 延迟删除策略
	key := fmt.Sprintf("user:friends:%d", f.UserID)
	err := dao.UpdataUserWithDelayDoubleDelete(key, func() *cerror.CodeError {
		targetUser := model.Users{}
		err := dao.DB.Table("users").Where("username = ?", req.TargetUsername).First(&targetUser).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return cerror.NewCodeError(4444, "目标用户不存在")
			}
			return cerror.NewCodeError(4444, err.Error())
		}
		// 在好友表中存入自己的好友信息
		friendTable := model.UserFriends{}
		err = dao.DB.Table(friendTable.GetTable()).Where("user_id = ? AND friend_id = ?", f.UserID, targetUser.Id).First(&friendTable).Error
		if err == nil {
			if friendTable.Status == 1 {
				return cerror.NewCodeError(4444, "已经是好友")
			}
		}
		// 插入一条好友关系记录
		friendTable = model.UserFriends{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			DeletedAt: nil,
			UserID:    int(f.UserID),
			FriendID:  int(targetUser.Id),
			Status:    0,
		}
		err = dao.DB.Table(friendTable.GetTable()).Create(&friendTable).Error
		if err != nil {
			return cerror.NewCodeError(4444, err.Error())
		}
		// 发送好友请求
		if targetUser.Status == 1 {
			// 目标用户在线， 发送好友请求
			f.sendFriendRequest(targetUser.Uuid, req.Content)
		}
		return nil
	})
	return err
}

// ✅
func (f *FriendService) sendFriendRequest(targetUUID string, content string) error {
	// 向对方发起好友通知
	for _, worker := range ServerInstance.WorkerHouse.Workers {
		if client := worker.GetClient(targetUUID); client != nil {
			// 编码成protoc
			notification := protocol.Message{
				FromUsername: f.Username,
				From:         f.UUID,
				To:           targetUUID,
				Content:      content,
				ContentType:  8,
				MessageType:  1,
			}
			notiByte, err := proto.Marshal(&notification)
			if err != nil {
				return err
			}
			client.Send <- notiByte
		}
	}
	return nil
}

// ✅
func (f *FriendService) GetFriendList() ([]dto.FriendInfo, *cerror.CodeError) {
	// 增加Redis
	// user:friends:{user_id}
	ctx := context.Background()
	key := fmt.Sprintf("user:friends:%d", f.UserID)
	exists := dao.REDIS.Exists(ctx, key).Val()
	if exists == 0 {
		logger.Info("缓存未命中")
		friendIDs := []int{}
		dao.DB.Table("user_friends").Where("user_id = ? AND status = 1", f.UserID).Pluck("friend_id", &friendIDs)
		if len(friendIDs) == 0 {
			return []dto.FriendInfo{}, cerror.NewCodeError(4444, "没有好友")
		}
		// 根据好友ID查询好友信息
		friendList := []dto.FriendInfo{}
		err := dao.DB.Table("users").Where("id IN (?)", friendIDs).Find(&friendList).Error
		if err != nil {
			return []dto.FriendInfo{}, cerror.NewCodeError(4444, err.Error())
		}
		// 序列化每个好友信息并存储
		for _, friend := range friendList {
			friendJSON, err2 := json.Marshal(friend)
			if err2 != nil {
				return []dto.FriendInfo{}, cerror.NewCodeError(4444, err2.Error())
			}
			dao.REDIS.LPush(ctx, key, friendJSON)
		}
		// 设置过期时间（30分钟）
		dao.REDIS.Expire(ctx, key, 30*time.Minute)
		// 增加一个机器人助手
		key = fmt.Sprintf("user:%s:bot", f.UUID)
		botUUID, err := dao.REDIS.Get(ctx, key).Result()
		friendList = append(friendList, dto.FriendInfo{
			UUID:     botUUID,
			Username: "机器人助手",
			Email:    "bot@example.com",
			Nickname: "机器人助手",
			Avatar:   "https://example.com/bot-avatar.jpg",
		})
		if err != nil {
			return []dto.FriendInfo{}, cerror.NewCodeError(4444, err.Error())
		}
		return friendList, nil
	}
	logger.Info("缓存命中")
	// 获取所有好友JSON字符串
	friendJSONList, err := dao.REDIS.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, cerror.NewCodeError(4444, err.Error())
	}
	// 反序列化为结构体切片
	friendList := make([]dto.FriendInfo, 0, len(friendJSONList))
	for _, friendJSON := range friendJSONList {
		var friend dto.FriendInfo
		if err := json.Unmarshal([]byte(friendJSON), &friend); err != nil {
			return nil, cerror.NewCodeError(4444, err.Error())
		}
		friendList = append(friendList, friend)
	}

	// 增加一个机器人助手
	key = fmt.Sprintf("user:%s:bot", f.UUID)
	botUUID, _ := dao.REDIS.Get(ctx, key).Result()
	friendList = append(friendList, dto.FriendInfo{
		UUID:     botUUID,
		Username: "机器人助手",
		Email:    "bot@example.com",
		Nickname: "机器人助手",
		Avatar:   "https://example.com/bot-avatar.jpg",
	})

	return friendList, nil
}

// ✅
func (f *FriendService) HandleFriendRequest(dto *dto.HandleFriendRequest) *cerror.CodeError {
	key := fmt.Sprintf("user:friends:%d", f.UserID)
	err := dao.UpdataUserWithDelayDoubleDelete(key, func() *cerror.CodeError {
		// 更新加好友的信息
		targetUser := model.Users{}
		err := dao.DB.Table(targetUser.GetTable()).Where("username = ?", dto.TargetUsername).First(&targetUser).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return cerror.NewCodeError(8988, err.Error())
		}
		friendTable := model.UserFriends{}
		err = dao.DB.Table(friendTable.GetTable()).Where("user_id = ? AND friend_id = ?", f.UserID, targetUser.Id).First(&friendTable).Error
		// 存在这条记录
		if friendTable.Id != 0 {
			return cerror.NewCodeError(8989, "好友已存在")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 创建一个好友Record
			newFriendRecord := model.UserFriends{
				UserID:    int(f.UserID),
				FriendID:  targetUser.Id,
				Status:    1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				DeletedAt: nil,
			}
			err = dao.DB.Table(friendTable.GetTable()).Create(&newFriendRecord).Error
			if err != nil {
				return cerror.NewCodeError(8989, err.Error())
			}
			// 将好友列表中对方的状态改为1
			err = dao.DB.Table(friendTable.GetTable()).Where("user_id = ? AND friend_id = ?", targetUser.Id, f.UserID).Update("status", 1).Error
			if err != nil {
				return cerror.NewCodeError(8989, err.Error())
			}
		} else {
			return cerror.NewCodeError(8989, err.Error())
		}
		return nil
	})
	return err
}
