package logic

import (
	"context"

	"chatLion/api/friend/internal/svc"
	"chatLion/api/friend/internal/types"
	"chatLion/rpc/friend/friend"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddFriendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddFriendLogic {
	return &AddFriendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddFriendLogic) AddFriend(req *types.AddFriendRequest) (resp *types.AddFriendResponse, err error) {
	// todo: add your logic here and delete this line
	addResp, err := l.svcCtx.FriendRPC.AddFriend(l.ctx, &friend.AddFriendRequest{
		AdderEmail:  req.AdderEmail,
		TargetEmail: req.TargetEmail,
		Content:     req.Content,
	})
	return &types.AddFriendResponse{
		AddMessage: addResp.AddMessage,
	}, err
}
