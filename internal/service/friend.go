package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/internal/dto"
	"cchat/pkg/cerror"
	"cchat/pkg/protocol"
	"errors"
	"time"

	"github.com/gogo/protobuf/proto"
	"gorm.io/gorm"
)

// ✅
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

// ✅
func AddFriend(req *dto.AddFriendRequest, senderId int, senderUuid string, senderName string) *cerror.CodeError {
	// 离线用户， 不经过websocket
	// 在线用户， 经过websocket
	// 根据信息查询uuid
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
	err = dao.DB.Table(friendTable.GetTable()).Where("user_id = ? AND friend_id = ?", senderId, targetUser.Id).First(&friendTable).Error
	if err == nil {
		if friendTable.Status == 1 {
			return cerror.NewCodeError(4444, "已经是好友")
		} else {
			// 好友请求已发送
			return cerror.NewCodeError(4444, "好友请求已发送")
		}
	}
	// 插入一条好友关系记录
	friendTable = model.UserFriends{
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
		DeleteAt: nil,
		UserID:   senderId,
		FriendID: int(targetUser.Id),
		Status:   0,
	}
	err = dao.DB.Table(friendTable.GetTable()).Create(&friendTable).Error
	if err != nil {
		return cerror.NewCodeError(4444, err.Error())
	}
	// 发送好友请求
	if targetUser.Status == 1 {
		// 目标用户在线， 发送好友请求
		SendFriendRequest(targetUser.Uuid, senderId, senderUuid, senderName, req.Content, 8)
	}
	return nil
}

// ✅
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

// ✅
func GetFriendList(userID int) ([]dto.FriendInfo, *cerror.CodeError) {
	friendIDs := []int{}
	dao.DB.Table("user_friends").Where("user_id = ? AND status = 1", userID).Pluck("friend_id", &friendIDs)
	if len(friendIDs) == 0 {
		return []dto.FriendInfo{}, cerror.NewCodeError(4444, "没有好友")
	}
	// 根据好友ID查询好友信息
	friendList := []dto.FriendInfo{}
	err := dao.DB.Table("users").Where("id IN (?)", friendIDs).Find(&friendList).Error
	if err != nil {
		return []dto.FriendInfo{}, cerror.NewCodeError(4444, err.Error())
	}
	return friendList, nil
}

// ✅
func HandleFriendRequest(dto *dto.HandleFriendRequest, userId int) *cerror.CodeError {
	// 更新加好友的信息
	targetUser := model.Users{}
	err := dao.DB.Table(targetUser.GetTable()).Where("username = ?", dto.TargetUsername).First(&targetUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return cerror.NewCodeError(8988, err.Error())
	}
	friendTable := model.UserFriends{}
	err = dao.DB.Table(friendTable.GetTable()).Where("user_id = ? AND friend_id = ?", userId, targetUser.Id).First(&friendTable).Error
	// 存在这条记录
	if friendTable.Id != 0 {
		return cerror.NewCodeError(8989, "好友已存在")
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建一个好友Record
		newFriendRecord := model.UserFriends{
			UserID:   userId,
			FriendID: targetUser.Id,
			Status:   1,
			CreateAt: time.Now(),
			UpdateAt: time.Now(),
			DeleteAt: nil,
		}
		err = dao.DB.Table(friendTable.GetTable()).Create(&newFriendRecord).Error
		if err != nil {
			return cerror.NewCodeError(8989, err.Error())
		}
		// 将好友列表中对方的状态改为1
		err = dao.DB.Table(friendTable.GetTable()).Where("user_id = ? AND friend_id = ?", targetUser.Id, userId).Update("status", 1).Error
		if err != nil {
			return cerror.NewCodeError(8989, err.Error())
		}
	} else {
		return cerror.NewCodeError(8989, err.Error())

	}
	return nil
}
