package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/internal/dto"
	"cchat/pkg/cerror"
	"cchat/pkg/logger"
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

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
	// 增加Redis缓存
	// 缓存key = user:info:uuid
	ctx := context.Background()
	key := fmt.Sprintf("user:info:%s", uuid)
	existes := dao.REDIS.Exists(ctx, key).Val()
	if existes == 0 {
		logger.Info("缓存不存在，从数据库查询", zap.String("key", key))
		// 缓存不存在，从数据库查询
		var user model.Users
		// 缓存未命中，从数据库中查询用户信息
		err := dao.DB.Table(user.GetTable()).Where("uuid = ?", uuid).First(&user).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, cerror.NewCodeError(1023, "在数据中查找用户不存在")
			}
			return nil, cerror.NewCodeError(1024, "查找失败: "+err.Error())
		}
		// 缓存用户信息
		// 使用Hash存储用户信息
		userMap := map[string]interface{}{
			"id":       user.Id,
			"uuid":     user.Uuid,
			"username": user.Username,
			"nickname": user.Nickname,
			"email":    user.Email,
			"avatar":   user.Avatar,
			"status":   user.Status,
		}

		// 设置Hash和过期时间
		dao.REDIS.HMSet(ctx, key, userMap)
		dao.REDIS.Expire(ctx, key, 30*time.Minute)
		// 设置映射关系
		dao.REDIS.Set(ctx, fmt.Sprintf("user:uuid:%s", user.Username), user.Uuid, 30*time.Minute)
		dao.REDIS.Set(ctx, fmt.Sprintf("user:uuid:email:%s", user.Email), user.Uuid, 30*time.Minute)
		// 构建返回数据
		userInfo := &dto.UserInfo{
			Email:    user.Email,
			Username: user.Username,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			UUID:     user.Uuid,
		}
		return userInfo, nil
	}

	// 获取Hash所有字段
	userMap := dao.REDIS.HGetAll(ctx, key).Val()
	if len(userMap) == 0 {
		return nil, cerror.NewCodeError(1024, "缓存数据为空")
	}
	// 从缓存中获取用户信息
	// 转换为DTO
	userInfo := &dto.UserInfo{
		Email:    userMap["email"],
		Username: userMap["username"],
		Nickname: userMap["nickname"],
		Avatar:   userMap["avatar"],
		UUID:     userMap["uuid"],
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
