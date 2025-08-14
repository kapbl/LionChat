package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/internal/dto"
	"cchat/pkg/cerror"
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

// GetUserIDByUUID 根据UUID获取用户ID
func GetUserIDByUUID(uuid string) (int, error) {
	if uuid == "" {
		return 0, errors.New("UUID不能为空")
	}

	// 查询用户ID
	var user model.Users
	err := dao.DB.Table(user.GetTable()).Select("id").Where("uuid = ?", uuid).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("用户不存在")
		}
		return 0, errors.New("查询用户ID失败: " + err.Error())
	}

	return int(user.Id), nil
}

// GetUserInfor 根据UUID获取用户信息
func GetUserInfor(uuid string) (*dto.UserInfo, *cerror.CodeError) {
	if uuid == "" {
		return nil, cerror.NewCodeError(1021, "UUID不能为空")
	}

	// 查询用户信息
	var user model.Users
	err := dao.DB.Table(user.GetTable()).Where("uuid = ?", uuid).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cerror.NewCodeError(1023, "在数据中查找用户不存在")
		}
		return nil, cerror.NewCodeError(1024, "查找失败: "+err.Error())
	}

	// 构建返回数据
	userInfo := &dto.UserInfo{
		Email:    user.Email,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
	}

	return userInfo, nil
}

// UpdateUserInfor 更新用户信息（不包括密码）
func UpdateUserInfor(uuid string, req *dto.UpdateUserReq) *cerror.CodeError {
	// 首先检查用户是否存在
	existingUser := model.Users{}
	err := dao.DB.Table(existingUser.GetTable()).Where("uuid = ?", uuid).First(&existingUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return cerror.NewCodeError(1023, "用户不存在")
		}
		return cerror.NewCodeError(1024, "查询用户失败: "+err.Error())
	}
	// 构建更新数据
	updateData := make(map[string]interface{})
	updateData["update_at"] = time.Now()

	// 只更新非空字段
	if req.Nickname != "" {
		updateData["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		updateData["avatar"] = req.Avatar
	}

	// 如果没有要更新的字段
	if len(updateData) == 1 { // 只有update_at字段
		return cerror.NewCodeError(1025, "没有要更新的字段")
	}
	// 执行更新
	err = dao.DB.Table(existingUser.GetTable()).Where("uuid = ?", uuid).Updates(updateData).Error
	if err != nil {
		return cerror.NewCodeError(1026, "更新用户信息失败: "+err.Error())
	}
	return nil
}
