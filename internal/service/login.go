package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/internal/dto"
	"cchat/pkg/hash"
	"cchat/pkg/token"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

func Login(req *dto.LoginReq) (dto.LoginResp, error) {
	// 查询这个用户
	currentUser := model.Users{}
	err := dao.DB.Table(currentUser.GetTable()).Where("username = ?", req.Username).Find(&currentUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.LoginResp{}, errors.New("用户不存在")
	}
	// 验证密码
	if !hash.VerifyPassword(currentUser.Password, req.Password) {
		return dto.LoginResp{}, errors.New("用户名或密码错误")
	}
	// 检查用户是否已经登录
	_, err = dao.REDIS.Get(context.Background(), "user:online:"+currentUser.Uuid).Result()
	if err == nil {
		return dto.LoginResp{}, errors.New("用户已登录")
	}
	// 登录成功后，将用户id存储到redis中
	// 格式：user:online:uuid -> id
	// 过期时间：5分钟
	dao.REDIS.Set(context.Background(), "user:online:"+currentUser.Uuid, currentUser.Id, 5*time.Minute)
	token, err := token.GEnToken(&currentUser)
	if err != nil {
		return dto.LoginResp{}, errors.New("生成token失败")
	}
	return dto.LoginResp{
		Code: 200,
		Msg:  "登录成功",
		Data: dto.LoginData{
			Token: token,
			UserInfo: dto.UserInfo{
				UUID:     currentUser.Uuid,
				Username: currentUser.Username,
				Nickname: currentUser.Nickname,
			},
		},
	}, nil
}
