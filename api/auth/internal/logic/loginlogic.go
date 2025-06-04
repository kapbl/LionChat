package logic

import (
	"context"
	"errors"

	"chatLion/api/auth/internal/svc"
	"chatLion/api/auth/internal/types"
	myjwt "chatLion/jwt"
	"chatLion/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	var user_model model.UserModel
	// !查看用户是否存在
	err = l.svcCtx.Db.Take(&user_model, "email = ?", req.Email).Error
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// !如果用户存在--继续验证密码是否正确
	if req.Password != user_model.Password {
		return nil, errors.New("密码不对")
	}
	//!如果密码正确-则生成jwt
	jwt := myjwt.JWTEncoder("chatLion", user_model.Email)

	// !如果密码正确-则返回可以进入主界面的消息
	resp = &types.LoginResponse{
		JWT: jwt,
	}
	return resp, nil
}
