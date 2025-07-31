package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

// UserInfoResp 用户信息响应结构体
type UserInfoResp struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
}

// UpdateUserReq 用户信息更新请求结构体
type UpdateUserReq struct {
	Username string `json:"username"` // 用户名
	Nickname string `json:"nickname"` // 昵称
	Email    string `json:"email"`    // 邮箱
	Avatar   string `json:"avatar"`   // 头像
}

// GetUserInfor 根据UUID获取用户信息
func GetUserInfor(uuid string) (*UserInfoResp, error) {
	if uuid == "" {
		return nil, errors.New("UUID不能为空")
	}

	// 查询用户信息
	var user model.Users
	err := dao.DB.Table(user.GetTable()).Where("uuid = ?", uuid).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, errors.New("查询用户信息失败: " + err.Error())
	}

	// 构建返回数据
	userInfo := &UserInfoResp{
		UUID:     user.Uuid,
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
		Avatar:   user.Avatar,
	}

	return userInfo, nil
}


// UpdateUserInfor 更新用户信息（不包括密码）
func UpdateUserInfor(uuid string, req *UpdateUserReq) error {
	if uuid == "" {
		return errors.New("UUID不能为空")
	}

	if req == nil {
		return errors.New("更新请求不能为空")
	}

	// 首先检查用户是否存在
	var existingUser model.Users
	err := dao.DB.Table(existingUser.GetTable()).Where("uuid = ?", uuid).First(&existingUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("用户不存在")
		}
		return errors.New("查询用户失败: " + err.Error())
	}

	// 检查用户名是否已被其他用户使用（如果要更新用户名）
	if req.Username != "" && req.Username != existingUser.Username {
		var count int64
		err = dao.DB.Table(existingUser.GetTable()).Where("username = ? AND uuid != ?", req.Username, uuid).Count(&count).Error
		if err != nil {
			return errors.New("检查用户名失败: " + err.Error())
		}
		if count > 0 {
			return errors.New("用户名已存在")
		}
	}

	// 构建更新数据
	updateData := make(map[string]interface{})
	updateData["update_at"] = time.Now()

	// 只更新非空字段
	if req.Username != "" {
		updateData["username"] = req.Username
	}
	if req.Nickname != "" {
		updateData["nickname"] = req.Nickname
	}
	if req.Email != "" {
		updateData["email"] = req.Email
	}
	if req.Avatar != "" {
		updateData["avatar"] = req.Avatar
	}

	// 如果没有要更新的字段
	if len(updateData) == 1 { // 只有update_at字段
		return errors.New("没有要更新的字段")
	}

	// 执行更新
	err = dao.DB.Table(existingUser.GetTable()).Where("uuid = ?", uuid).Updates(updateData).Error
	if err != nil {
		return errors.New("更新用户信息失败: " + err.Error())
	}

	return nil
}