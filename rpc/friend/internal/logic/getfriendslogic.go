package logic

import (
	"context"

	"chatLion/model"
	"chatLion/rpc/friend/friend"
	"chatLion/rpc/friend/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFriendsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFriendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendsLogic {
	return &GetFriendsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFriendsLogic) GetFriends(in *friend.GetFriendsListRequest) (*friend.GetFriendsListResponse, error) {
	// todo: add your logic here and delete this line
	// 现在数据库中查询该用户的所有好友们
	var currentUser model.UserModel
	if err := l.svcCtx.Db.First(&currentUser, "email = ?", in.UserEmail).Error; err != nil {
		return &friend.GetFriendsListResponse{
			Friends: nil,
		}, err
	}

	// 查询currentUser的所有好友
	var friends []model.UserFriend
	// 查询user_id = 5的所有好友
	l.svcCtx.Db.Where("user_id = ?", currentUser.ID).Find(&friends)
	friendAns := make([]*friend.FriendInfo, len(friends))
	for i, f := range friends {
		var curFriendUser model.UserModel
		l.svcCtx.Db.First(&curFriendUser, "id = ?", f.FriendID)
		friendAns[i] = &friend.FriendInfo{
			Email:    curFriendUser.Email,
			Nickname: curFriendUser.Nickname,
		}
	}

	return &friend.GetFriendsListResponse{
		Friends: friendAns,
	}, nil
}
