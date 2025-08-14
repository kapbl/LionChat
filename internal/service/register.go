package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/internal/dto"
	"cchat/pkg/cerror"
	"cchat/pkg/hash"
	"cchat/pkg/token"
	"time"
)

func Register(req *dto.RegisterRequest) *cerror.CodeError {

	searchUser := model.Users{}
	// 注册的条件：邮箱和用户名必须唯一
	dao.DB.Table("users").Where("username = ? or email = ?", req.Username, req.Email).First(&searchUser)
	if searchUser.Id != 0 {
		return &cerror.CodeError{
			Code: 1005,
			Msg:  "用户名或邮箱已存在",
		}
	}
	// 加密密码
	hashedPassword, err := hash.HashPassword(req.Password)
	if err != nil {
		return &cerror.CodeError{
			Code: 1007,
			Msg:  "密码加密失败",
		}
	}
	// 初始化用户
	newUser := model.Users{
		Username: req.Username,
		Nickname: req.Nickname,
		Password: hashedPassword,
		Email:    req.Email,
		Uuid:     token.GenUUID(req.Username),
		Status:   0,
		CreateAt: time.Now(),
		UpdateAt: time.Now(),
		DeleteAt: nil,
	}
	// 保存用户
	err = dao.DB.Table("users").Create(&newUser).Error
	if err != nil {
		return &cerror.CodeError{
			Code: 1009,
			Msg:  "数据库写入失败",
		}
	}
	return nil
}
