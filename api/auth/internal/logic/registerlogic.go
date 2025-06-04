package logic

import (
	"context"
	"errors"

	"chatLion/api/auth/internal/svc"
	"chatLion/api/auth/internal/types"
	"chatLion/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.RegisterResponse, err error) {
	//!查看用户是否存在
	var user_model_queue model.UserModel
	err = l.svcCtx.Db.Take(&user_model_queue, "email =?", req.Email).Error
	if err == nil {
		return nil, errors.New("用户已存在")
	}
	// !如果用户不存在-则注册
	var new_user model.UserModel
	new_user.Email = req.Email
	new_user.Password = req.Password
	err = l.svcCtx.Db.Create(&new_user).Error
	if err != nil {
		return nil, errors.New("注册失败")
	}
	resp = &types.RegisterResponse{
		Message: "注册成功",
	}
	return resp, nil
}
