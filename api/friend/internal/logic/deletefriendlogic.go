package logic

import (
	"context"

	"chatLion/api/friend/internal/svc"
	"chatLion/api/friend/internal/types"
	"chatLion/rpc/friend/friendclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteFriendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFriendLogic {
	return &DeleteFriendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteFriendLogic) DeleteFriend(req *types.DeleteFriendRequest) (resp *types.DeleteFriendResponse, err error) {
	// todo: add your logic here and delete this line
	deleteResp, err := l.svcCtx.FriendRPC.DeleteFriend(l.ctx, &friendclient.DeleteFriendRequest{
		DeleteEmail: req.DeleteEmail,
		TargetEmail: req.TargetEmail,
	})
	return &types.DeleteFriendResponse{
		DeleteMessage: deleteResp.DeleteMessage,
	}, err
}
