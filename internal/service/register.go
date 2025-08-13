package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/internal/dto"
	"cchat/pkg/hash"
	"cchat/pkg/token"
	"errors"
	"time"
)

func Register(req *dto.RegisterReq) (*dto.RegisterResp, error) {
	var searchUser model.Users
	err := dao.DB.Table("users").Where("username = ?", req.Username).First(&searchUser).Error
	if err == nil {
		return nil, errors.New("用户名已存在")
	}
	// 加密密码
	hashedPassword, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	searchUser.Username = req.Username
	searchUser.Nickname = req.Nickname
	searchUser.Password = hashedPassword
	// 生成唯一的uuid
	uuid := token.GenUUID(req.Username)
	searchUser.Uuid = uuid
	searchUser.CreateAt = time.Now()
	searchUser.UpdateAt = time.Now()
	searchUser.DeleteAt = nil
	// 保存用户
	err = dao.DB.Table("users").Create(&searchUser).Error
	if err != nil {
		return nil, errors.New("注册失败")
	}
	return &dto.RegisterResp{
		Code: 200,
		Msg:  uuid,
	}, nil
}
