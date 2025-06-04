package logic

import (
	"context"

	"chatLion/model"
	"chatLion/rpc/user/internal/svc"
	"chatLion/rpc/user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type ModifyUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewModifyUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ModifyUserLogic {
	return &ModifyUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ModifyUserLogic) ModifyUser(in *user.ModifyUserRequest) (*user.ModifyUserResponse, error) {
	// todo: add your logic here and delete this line
	var currentUser model.UserModel
	if err := l.svcCtx.Db.First(&currentUser, "email = ?", in.Email).Error; err != nil {
		return &user.ModifyUserResponse{
			ModifyMessage: "用户不存在",
		}, err
	}
	err := l.svcCtx.Db.Model(&currentUser).Update("nickname", in.NewUserNickname).Error
	if err != nil {
		return &user.ModifyUserResponse{
			ModifyMessage: "更新用户信息失败",
		}, err
	}
	return &user.ModifyUserResponse{
		ModifyMessage: "更新用户信息成功",
	}, nil
}
