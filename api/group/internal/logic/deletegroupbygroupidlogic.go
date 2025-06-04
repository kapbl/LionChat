package logic

import (
	"context"

	"chatLion/api/group/internal/svc"
	"chatLion/api/group/internal/types"
	"chatLion/rpc/group/group"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteGroupByGroupIDLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteGroupByGroupIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteGroupByGroupIDLogic {
	return &DeleteGroupByGroupIDLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteGroupByGroupIDLogic) DeleteGroupByGroupID(req *types.DeleteGroupRequest) (resp *types.DeleteGroupResponse, err error) {
	// todo: add your logic here and delete this line
	deleteResp, err := l.svcCtx.GroupRPC.DeleteGroup(l.ctx, &group.DeleteGroupRequest{
		DeleteGroupName: req.DeleteGroupName,
		UserName:        req.UserName,
	})
	return &types.DeleteGroupResponse{
		DeleteMessage: deleteResp.DeleteMessage,
	}, err
}
