package logic

import (
	"context"
	"log"

	"chatLion/model"
	"chatLion/rpc/friend/friend"
	"chatLion/rpc/friend/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeleteFriendLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFriendLogic {
	return &DeleteFriendLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteFriendLogic) DeleteFriend(in *friend.DeleteFriendRequest) (*friend.DeleteFriendResponse, error) {
	var currentUser model.UserModel
	l.svcCtx.Db.First(&currentUser, "email = ?", in.DeleteEmail)
	log.Println(currentUser.ID)

	var targetUser model.UserModel
	l.svcCtx.Db.First(&targetUser, "email = ?", in.TargetEmail)
	log.Println(targetUser.ID)

	// 开始尝试删除好友
	err := DeleteFriend(l.svcCtx.Db, currentUser.ID, targetUser.ID)
	if err != nil{
		return &friend.DeleteFriendResponse{
			DeleteMessage: "删除好友失败",
		},err
	}
	return &friend.DeleteFriendResponse{
		DeleteMessage: "删除好友成功",
	}, nil
}

func DeleteFriend(db *gorm.DB, userID, friendID uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 删除正向关系 (A→B)
		if err := tx.Where("user_id = ? AND friend_id = ?", userID, friendID).
			Delete(&model.UserFriend{}).Error; err != nil {
			return err
		}
		return nil
	})
}
