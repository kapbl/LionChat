package logic

import (
	"context"

	"chatLion/api/group/internal/svc"
	"chatLion/api/group/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMembersByGroupIDLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMembersByGroupIDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMembersByGroupIDLogic {
	return &GetMembersByGroupIDLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMembersByGroupIDLogic) GetMembersByGroupID(req *types.GetMemberRequest) (resp *types.GetMemberResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
