package logic

import (
	"context"

	"chatLion/rpc/friend/friend"
	"chatLion/rpc/friend/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
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

	return &friend.AddFriendResponse{}, nil
}
