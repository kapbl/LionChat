package logic

import (
	"context"
	"errors"

	"chatLion/api/friend/internal/svc"
	"chatLion/api/friend/internal/types"
	"chatLion/rpc/friend/friend"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFriendsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFriendsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendsLogic {
	return &GetFriendsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFriendsLogic) GetFriends(req *types.GetFriendsRequest) (resp *types.GetFriendsResponse, err error) {
	getResp, err := l.svcCtx.FriendRPC.GetFriends(l.ctx, &friend.GetFriendsListRequest{
		UserEmail: req.UserEmail,
	})
	if err != nil {
		return &types.GetFriendsResponse{
			Friends: nil,
		}, errors.New("error get friends")
	}
	if len(getResp.Friends) == 0 {
		return &types.GetFriendsResponse{
			Friends: nil,
		}, nil
	}
	friends := make([]types.FriendInfor, len(getResp.Friends))
	for i, friend := range getResp.Friends {
		friends[i] = types.FriendInfor{Nickname: friend.Nickname, Email: friend.Email}
	}
	return &types.GetFriendsResponse{
		Friends: friends,
	}, nil
}
