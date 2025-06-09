package logic

import (
	"context"

	"chatLion/model"
	"chatLion/rpc/user/internal/svc"
	"chatLion/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyUserAvatorLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewModifyUserAvatorLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyUserAvatorLogic {
	return &ModifyUserAvatorLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ModifyUserAvatorLogic) ModifyUserAvator(in *user.MoifyUserAvatorRequest) (*user.MoifyUserAvatorResponse, error) {
	// todo: add your logic here and delete this line
	var currentUser model.UserModel
	if err := l.svcCtx.Db.First(&currentUser, "email = ?", in.Email).Error; err != nil {
		return &user.MoifyUserAvatorResponse{
			ModifyMessage: "非法用户",
		}, err
	}
	err := l.svcCtx.Db.Update("avator", in.AvatorNum).Where("email =?", currentUser.Email).Error
	if err != nil {
		return &user.MoifyUserAvatorResponse{
			ModifyMessage: "修改失败",
		}, err
	}
	return &user.MoifyUserAvatorResponse{
		ModifyMessage: "修改成功",
	}, nil
}
