package logic

import (
	"context"

	"chatLion/api/user/internal/svc"
	"chatLion/api/user/internal/types"
	"chatLion/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewModifyUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyUserLogic {
	return &ModifyUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ModifyUserLogic) ModifyUser(req *types.ModifyUserRequest) (resp *types.ModifyUserResponse, err error) {
	// todo: add your logic here and delete this line
	modifyResp, err := l.svcCtx.UserRPC.ModifyUser(l.ctx, &user.ModifyUserRequest{
		Email:           req.Email,
		NewUserNickname: req.NewUserName,
	})

	return &types.ModifyUserResponse{
		ModifyMessage: modifyResp.ModifyMessage,
	}, err
}
