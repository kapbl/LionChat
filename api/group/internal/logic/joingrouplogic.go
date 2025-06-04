package logic

import (
	"context"

	"chatLion/api/group/internal/svc"
	"chatLion/api/group/internal/types"
	"chatLion/rpc/group/groupclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type JoinGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewJoinGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *JoinGroupLogic {
	return &JoinGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *JoinGroupLogic) JoinGroup(req *types.JoinGroupRequest) (resp *types.JoinGroupResponse, err error) {
	// todo: add your logic here and delete this line
	joinRequest := &groupclient.JoinGroupRequest{
		JoinGroupName: req.JoinGroupName,
		JoinerName:    req.JoinerName,
	}
	joinResponse, err := l.svcCtx.GroupRPC.JoinGroup(l.ctx, joinRequest)
	if err != nil {
		return &types.JoinGroupResponse{
			JoinMessage: joinResponse.JoinMessage,
		}, err
	}
	return &types.JoinGroupResponse{
		JoinMessage: joinResponse.JoinMessage,
	}, nil
}
