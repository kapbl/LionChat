package logic

import (
	"context"

	"chatLion/api/group/internal/svc"
	"chatLion/api/group/internal/types"
	"chatLion/rpc/group/groupclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGroupLogic {
	return &CreateGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateGroupLogic) CreateGroup(req *types.CreateGroupRequest) (resp *types.CreateGroupResponse, err error) {
	// 在数据库中查询新组名是否存在
	createGroupRequest := new(groupclient.CreateGroupRequest)
	createGroupRequest.CreateGroupName = req.CreateGroupName
	createGroupRequest.CreateGroupDescription = req.CreateGroupDescription
	createGroupRequest.CreateOwnerName = req.CreateGroupOwnerName

	createResponse, err := l.svcCtx.GroupRPC.CreateGroup(l.ctx, createGroupRequest)

	if err != nil {
		return &types.CreateGroupResponse{
			CreateMessage: "创建组失败",
		}, err
	}

	return &types.CreateGroupResponse{
		CreateMessage: createResponse.CreateMessage,
	}, nil
}
