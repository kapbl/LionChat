package logic

import (
	"context"
	"errors"

	"chatLion/model"
	"chatLion/rpc/friend/friend"
	"chatLion/rpc/friend/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type AddFriendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddFriendLogic {
	return &AddFriendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AddFriendLogic) AddFriend(in *friend.AddFriendRequest) (*friend.AddFriendResponse, error) {
	// todo: add your logic here and delete this line
	var targetUser model.UserModel
	err := l.svcCtx.Db.First(&targetUser, "email = ?", in.TargetEmail).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &friend.AddFriendResponse{
				AddMessage: "未查询到目标",
			}, err
		} else {
			return &friend.AddFriendResponse{
				AddMessage: "添加好友失败",
			}, err
		}
	}
	// 判断是否已经是好友了
	var mine model.UserModel
	err = l.svcCtx.Db.First(&mine, "email = ?", in.AdderEmail).Error
	if err != nil {
		return &friend.AddFriendResponse{
			AddMessage: "添加好友失败",
		}, err
	}
	for _, v := range mine.Friends {
		if v.ID == targetUser.ID {
			return &friend.AddFriendResponse{
				AddMessage: "对方已经是您的好友了",
			}, nil
		}
	}
	// 开始尝试添加好友
	userFriend := model.UserFriend{
		UserID:   mine.ID,
		FriendID: targetUser.ID,
	}
	err = l.svcCtx.Db.Create(&userFriend).Error
	if err != nil {
		return &friend.AddFriendResponse{
			AddMessage: "添加好友失败",
		}, err
	}

	return &friend.AddFriendResponse{
		AddMessage: "成功添加" + targetUser.Nickname + "为你的好友",
	}, err
}
