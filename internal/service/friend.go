package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/internal/dto"
	"cchat/pkg/cerror"
	"cchat/pkg/protocol"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
	"gorm.io/gorm"
)

func SearchClient(information string) (*dto.UserInfo, *cerror.CodeError) {
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

func AddSearchClientByUserName(req *dto.AddFriendReq, userId int, uuid string, userName string) (dto.AddFriendResp, error) {
	// 查询对方
	targetUser := model.Users{}
	err := dao.DB.Table(targetUser.GetTable()).Where("username = ?", req.TargetUserName).First(&targetUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.AddFriendResp{}, errors.New("目标用户不存在")
	}
	// 查询是否已经是好友 查询是否发送过请求
	friendTable := model.UserFriends{}
	err = dao.DB.Table(friendTable.GetTable()).Where("user_id = ? AND friend_id = ?", userId, targetUser.Id).First(&friendTable).Error
	if err == nil {
		SendFriendRequest(targetUser.Uuid, userId, uuid, userName, req.Content, 8)
		return dto.AddFriendResp{
			OriginUUID:     uuid,
			TargetUserName: targetUser.Username,
		}, nil
	}

	// 延迟双删
	err = dao.UpdataUserWithDelayDoubleDelete(uuid, func() error {
		// 查询自己的群组的最新版本号
		// 在数据库中添加这条加好友记录
		friendTable = model.UserFriends{
			CreateAt: time.Now(),
			UpdateAt: time.Now(),
			DeleteAt: nil,
			UserID:   userId,
			FriendID: int(targetUser.Id),
			Status:   0,
		}
		err = dao.DB.Table(friendTable.GetTable()).Create(&friendTable).Error
		dao.DB.Table("users").Where("id = ?", userId).
			Update("friend_version", gorm.Expr("friend_version + ?", 1))
		if err != nil {
			return nil
		}
		return nil
	})
	if err != nil {
		return dto.AddFriendResp{}, err
	}
	SendFriendRequest(targetUser.Uuid, userId, uuid, userName, req.Content, 8)
	return dto.AddFriendResp{
		OriginUUID:     uuid,
		TargetUserName: targetUser.Username,
	}, nil
}

func SendFriendRequest(targetUUID string, userId int, uuid string, userName string, content string, contentType int) error {
	// 向对方发起好友通知
	for _, worker := range ServerInstance.WorkerHouse.Workers {
		if client := worker.GetClient(targetUUID); client != nil {
			// 编码成protoc
			notification := protocol.Message{
				FromUsername: userName,
				From:         uuid,
				To:           targetUUID,
				Content:      content,
				ContentType:  int32(contentType),
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

func GetFriendList(uuid string, userID int) ([]dto.FriendInfo, error) {
	var friendList []dto.FriendInfo
	// 先从Redis中间件
	ctx := context.Background()
	result, err := dao.REDIS.HGetAll(
		ctx,
		uuid,
	).Result()

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("redis timeout")
		}
	}

	// 缓存未命中
	if len(result) == 0 {
		// 直接通过 SQL 联表查询
		err := dao.DB.Table("user_friends AS uf").
			Select("u.uuid AS friend_uuid,u.username AS friend_name,u.avatar AS friend_avatar,u.nickname AS friend_nickname,uf.status").
			Joins("JOIN users AS u ON u.id = uf.friend_id").
			Where("uf.user_id = (SELECT id FROM users WHERE uuid = ?)", uuid).
			Where("uf.status = 1").
			Scan(&friendList).Error
		if err != nil {
			return []dto.FriendInfo{}, err
		}
		// 更新到redis中
		// 先查询版本号
		var user model.Users
		dao.DB.Table("users").Where("id = ?", userID).First(&user)
		for i := 0; i < len(friendList); i++ {
			friendList[i].Version = user.GroupVersion
		}
		// 再更新
		pipe := dao.REDIS.Pipeline()
		for _, f := range friendList {
			// 序列化结构体为JSON
			friendData, _ := json.Marshal(f)
			// 使用用户ID作为键，好友UUID作为字段
			pipe.HSet(ctx,
				uuid,
				f.FriendUUID,
				string(friendData),
			)
		}
		pipe.Expire(ctx, uuid, 10*time.Minute) // 设置 TTL
		_, _ = pipe.Exec(ctx)                  // 忽略缓存失败
		return friendList, nil
	}
	// 获自己的朋友版本号
	sentinel := model.Users{}
	dao.DB.Table("users").Where("id = ?", userID).First(&sentinel)
	// 如果在Redis中可以缓存到且与数据库的版本号一致
	for _, v := range result {
		var friend dto.FriendInfo
		if err := json.Unmarshal([]byte(v), &friend); err != nil {
			continue
		}
		if sentinel.FriendVersion != friend.Version {
			// 版本号不一致，需要重新从数据库查询
			err := dao.DB.Table("user_friends AS uf").
				Select("u.uuid AS friend_uuid,u.username AS friend_name,u.avatar AS friend_avatar,u.nickname AS friend_nickname,uf.status").
				Joins("JOIN users AS u ON u.id = uf.friend_id").
				Where("uf.user_id = (SELECT id FROM users WHERE uuid = ?)", uuid).
				Where("uf.status = 1").
				Scan(&friendList).Error
			if err != nil {
				return []dto.FriendInfo{}, err
			}
			// 先查询版本号
			var user model.Users
			dao.DB.Table("users").Where("id = ?", userID).First(&user)
			for i := 0; i < len(friendList); i++ {
				friendList[i].Version = user.FriendVersion
			}
			// 再更新到Redis
			pipe := dao.REDIS.Pipeline()
			for _, f := range friendList {
				// 序列化结构体为JSON
				friendData, _ := json.Marshal(f)
				// 使用用户ID作为键，好友UUID作为字段
				pipe.HSet(ctx,
					uuid,
					f.FriendUUID,
					string(friendData),
				)
			}
			pipe.Expire(ctx, uuid, 10*time.Minute) // 设置 TTL
			_, _ = pipe.Exec(ctx)                  // 忽略缓存失败
			return friendList, nil
		}
		friendList = append(friendList, friend)
	}
	return friendList, nil
}

func ReceiveFriendRequest(d *dto.HandleFriendRequest, userId int, uuid string, username string) error {
	if d.Status == 0 {
		SendFriendRequest(d.TargetUUID, userId, uuid, username, "不想加你", 9)
		return nil
	}
	// 查找对方的uuid是否正确
	targetUser := model.Users{}
	err := dao.DB.Table(targetUser.GetTable()).Where("uuid = ?", d.TargetUUID).First(&targetUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	// 如果uuid正确
	friendTable := model.UserFriends{}
	err = dao.DB.Table(friendTable.GetTable()).Where("user_id = ? AND friend_id = ?", userId, targetUser.Id).First(&friendTable).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 延迟双删
		err = dao.UpdataUserWithDelayDoubleDelete(uuid, func() error {
			// 创建一个好友Record
			newFriendRecord := model.UserFriends{
				CreateAt: time.Now(),
				UpdateAt: time.Now(),
				DeleteAt: nil,
				UserID:   userId,
				FriendID: int(targetUser.Id),
				Status:   1,
			}
			err = dao.DB.Table(friendTable.GetTable()).Create(&newFriendRecord).Error
			dao.DB.Table("users").Where("id = ?", userId).
				Update("group_version", gorm.Expr("group_version + ?", 1))
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
		SendFriendRequest(d.TargetUUID, userId, uuid, username, "好滴", 9)
	} else if err == nil {
		// 如果记录存在则变更状态
		err = dao.UpdataUserWithDelayDoubleDelete(uuid, func() error {
			dao.DB.Table(friendTable.GetTable()).Update("status", d.Status)
			dao.DB.Table("users").Where("id = ?", userId).
				Update("friend_version", gorm.Expr("friend_version + ?", 1))
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
		SendFriendRequest(d.TargetUUID, userId, uuid, username, "好滴", 9)
	}
	return nil
}

func HandleFriendRequest(dto *dto.HandleFriendRequest, userId int, uuid string, username string) error {
	// 更新加好友的信息
	// 查找对方的uuid是否正确
	targetUser := model.Users{}
	err := dao.DB.Table(targetUser.GetTable()).Where("uuid = ?", dto.TargetUUID).First(&targetUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	// 如果uuid正确
	friendTable := model.UserFriends{}
	err = dao.DB.Table(friendTable.GetTable()).Where("user_id = ? AND friend_id = ?", userId, targetUser.Id).First(&friendTable).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 延迟双删
		err = dao.UpdataUserWithDelayDoubleDelete(uuid, func() error {
			// 创建一个好友Record
			newFriendRecord := model.UserFriends{
				CreateAt: time.Now(),
				UpdateAt: time.Now(),
				DeleteAt: nil,
				UserID:   userId,
				FriendID: int(targetUser.Id),
				Status:   1,
			}
			err = dao.DB.Table(friendTable.GetTable()).Create(&newFriendRecord).Error
			dao.DB.Table("users").Where("id = ?", userId).
				Update("friend_version", gorm.Expr("friend_version + ?", 1))
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	} else if err == nil {
		// 如果记录存在则变更状态
		err = dao.UpdataUserWithDelayDoubleDelete(uuid, func() error {
			dao.DB.Table(friendTable.GetTable()).Where("user_id = ? AND friend_id = ?", userId, targetUser.Id).Update("status", dto.Status)
			dao.DB.Table("users").Where("id = ?", userId).
				Update("friend_version", gorm.Expr("friend_version + ?", 1))
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
