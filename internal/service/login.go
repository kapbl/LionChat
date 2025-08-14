package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/internal/dto"
	"cchat/pkg/cerror"
	"cchat/pkg/hash"
	"cchat/pkg/token"
	"errors"

	"gorm.io/gorm"
)

func Login(req *dto.LoginRequest) (string, *cerror.CodeError) {
	// 查询这个用户
	currentUser := model.Users{}
	err := dao.DB.Table(currentUser.GetTable()).Where("email = ? or username = ?", req.Account, req.Account).First(&currentUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", &cerror.CodeError{Code: 1003, Msg: "用户不存在"}
	}
	// 验证密码
	if !hash.VerifyPassword(currentUser.Password, req.Password) {
		return "", &cerror.CodeError{Code: 1003, Msg: "用户名或密码错误"}
	}
	accessToken, err := token.GEnToken(&currentUser)
	if err != nil {
		return "", &cerror.CodeError{Code: 1017, Msg: "生成token失败"}
	}
	return accessToken, nil
}
