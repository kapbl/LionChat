package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/internal/dto"
	"cchat/pkg/protocol"
	"errors"
	"time"

	"github.com/gogo/protobuf/proto"
	"gorm.io/gorm"
)

func SearchClientByUserName(username string) (dto.SearchFriendResp, error) {
	user := model.User{}

	err := dao.DB.Table(user.GetTable()).Where("username = ?", username).Find(&user).Error
	if err != nil {
		return dto.SearchFriendResp{}, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.SearchFriendResp{}, gorm.ErrRecordNotFound
	}

	resp := dto.SearchFriendResp{
		Username: user.Username,
		UUID:     user.Uuid,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
	}
	return resp, nil
}

func AddSearchClientByUserName(req *dto.AddFriendReq, userId int, uuid string, userName string) (dto.AddFriendResp, error) {
	// 查询对方
	targetUser := model.User{}
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

	// 在数据库中添加这条加好友记录
	friendTable = model.UserFriends{
		CreateAt: time.Now(),
		UpdateAt: nil,
		DeleteAt: 0,
		UserID:   userId,
		FriendID: int(targetUser.Id),
		Status:   0,
	}
	err = dao.DB.Table(friendTable.GetTable()).Create(&friendTable).Error
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
	if client := ServerInstance.GetClient(targetUUID); client != nil {
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
	return nil
}

func GetFriendList(userId int) ([]dto.FriendInfo, error) {
	// 左连接还是右连接
	var friendList []dto.FriendInfo
	// 直接通过 SQL 联表查询
	err := dao.DB.Table("user_friends AS uf").
		Select("u.uuid AS friend_uuid,u.username AS friend_name,u.avatar AS friend_avatar,u.nickname AS friend_nickname,uf.status").
		Joins("JOIN users AS u ON u.id = uf.friend_id").
		Where("uf.user_id = (SELECT id FROM users WHERE uuid = ?)", userId).
		Where("uf.status = 1").
		Scan(&friendList).Error
	if err != nil {
		return []dto.FriendInfo{}, err
	}
	return friendList, nil
}

func ReceiveFriendRequest(dto *dto.HandleFriendRequest, userId int, uuid string, username string) error {
	if dto.Status == 0 {
		SendFriendRequest(dto.TargetUUID, userId, uuid, username, "不想加你", 9)
		return nil
	}
	// 查找对方的uuid是否正确
	targetUser := model.User{}
	err := dao.DB.Table(targetUser.GetTable()).Where("uuid = ?", dto.TargetUUID).First(&targetUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	// 如果uuid正确
	friendTable := model.UserFriends{}
	err = dao.DB.Table(friendTable.GetTable()).Where("user_id = ? AND friend_id = ?", userId, targetUser.Id).First(&friendTable).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建一个好友Record
		newFriendRecord := model.UserFriends{
			CreateAt: time.Now(),
			UpdateAt: nil,
			DeleteAt: 0,
			UserID:   userId,
			FriendID: int(targetUser.Id),
			Status:   1,
		}
		err = dao.DB.Table(friendTable.GetTable()).Create(&newFriendRecord).Error
		if err != nil {
			return err
		}
		SendFriendRequest(dto.TargetUUID, userId, uuid, username, "好滴", 9)
	} else if err == nil {
		// 如果记录存在则变更状态
		dao.DB.Table(friendTable.GetTable()).Update("status", dto.Status)
		SendFriendRequest(dto.TargetUUID, userId, uuid, username, "好滴", 9)
	}
	return nil
}

func HandleFriendRequest(dto *dto.HandleFriendRequest, userId int, uuid string, username string) error {
	// 更新加好友的信息
	// 查找对方的uuid是否正确
	targetUser := model.User{}
	err := dao.DB.Table(targetUser.GetTable()).Where("uuid = ?", dto.TargetUUID).First(&targetUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	// 如果uuid正确
	friendTable := model.UserFriends{}
	err = dao.DB.Table(friendTable.GetTable()).Where("user_id = ? AND friend_id = ?", userId, targetUser.Id).First(&friendTable).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建一个好友Record
		newFriendRecord := model.UserFriends{
			CreateAt: time.Now(),
			UpdateAt: nil,
			DeleteAt: 0,
			UserID:   userId,
			FriendID: int(targetUser.Id),
			Status:   1,
		}
		err = dao.DB.Table(friendTable.GetTable()).Create(&newFriendRecord).Error
		if err != nil {
			return err
		}
	} else if err == nil {
		// 如果记录存在则变更状态

		dao.DB.Table(friendTable.GetTable()).Where("user_id = ? AND friend_id = ?", userId, targetUser.Id).Update("status", dto.Status)
	}
	return nil
}
