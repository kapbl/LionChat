package logic

import (
	"context"

	"chatLion/api/user/internal/svc"
	"chatLion/api/user/internal/types"
	"chatLion/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyUserAvatorLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewModifyUserAvatorLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyUserAvatorLogic {
	return &ModifyUserAvatorLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ModifyUserAvatorLogic) ModifyUserAvator(req *types.ModifyUserAvatorRequest) (resp *types.ModifyUserAvatorResponse, err error) {
	// todo: add your logic here and delete this line
	// var currentUser model.UserModel
	response, err := l.svcCtx.UserRPC.ModifyUserAvator(l.ctx, &user.MoifyUserAvatorRequest{
		Email:     req.Email,
		AvatorNum: int64(req.AvatorNum),
	})
	if err != nil {
		return &types.ModifyUserAvatorResponse{
			ModifyMessage: "修改成功",
		}, nil
	}
	return &types.ModifyUserAvatorResponse{
		ModifyMessage: response.ModifyMessage,
	}, nil
}
