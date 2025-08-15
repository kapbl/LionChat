package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/internal/dto"
	"cchat/pkg/logger"
	"errors"
	"time"

	"go.uber.org/zap"
)

func CreateMoment(moment *dto.MomentCreateReq, uuid string) (*dto.MomentCreateResp, error) {
	// 校验moment
	if moment == nil {
		return nil, errors.New("moment is nil")
	}
	// 查询用户
	var user model.Users
	if err := dao.DB.Where("uuid = ?", uuid).First(&user).Error; err != nil {
		return nil, err
	}
	// 在数据库中插入这条动态
	m := &model.Moment{
		UserID:    int64(user.Id),
		Content:   moment.Content,
		CreatedAt: time.Now(),
		DeletedAt: nil,
		UpdatedAt: time.Now(),
	}
	// 插入到自己的动态表中
	if err := dao.DB.Create(m).Error; err != nil {
		return nil, err
	}
	// 插入到自己的Timeline表中
	t := &model.Timeline{
		UserID:    int64(user.Id),
		MomentID:  m.ID,
		IsOwn:     true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: nil,
	}
	if err := dao.DB.Create(t).Error; err != nil {
		return nil, err
	}

	// 异步插入所有可见好友的Timeline表
	go func() {
		// 查询用户的所有好友
		var friends []*model.UserFriends
		if err := dao.DB.Where("user_id = ?", int64(user.Id)).Find(&friends).Error; err != nil {
			logger.Error("查询用户好友失败", zap.Error(err))
			return
		}
		// 插入到好友的Timeline表中
		for _, friend := range friends {
			t := &model.Timeline{
				UserID:    int64(friend.FriendID),
				MomentID:  m.ID,
				IsOwn:     false,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := dao.DB.Create(t).Error; err != nil {
				logger.Error("插入好友Timeline失败", zap.Error(err))
			}
		}

	}()
	// 返回动态ID
	return &dto.MomentCreateResp{ID: m.ID}, nil
}

// 获取用户的朋友圈动态列表
// 基于 timeline 表查询，按时间倒序展示
// GetCommentsByMomentID 根据动态ID查询评论列表
func GetCommentsByMomentID(momentID int64) ([]dto.CommentListResp, error) {
	// 查询评论记录，关联用户表获取用户名
	var results []struct {
		UserID     int64     `gorm:"column:user_id"`
		Username   string    `gorm:"column:username"`
		Content    string    `gorm:"column:content"`
		CreateTime time.Time `gorm:"column:created_at"`
	}

	// 从评论表查询指定动态的评论，按时间正序
	if err := dao.DB.Table("comment c").
		Select("c.user_id, u.username, c.content, c.created_at").
		Joins("JOIN users u ON c.user_id = u.id").
		Where("c.moment_id = ?", momentID).
		Order("c.created_at ASC").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	// 转换为响应格式
	var commentList []dto.CommentListResp
	for _, result := range results {
		commentList = append(commentList, dto.CommentListResp{
			UserID:     result.UserID,
			Username:   result.Username,
			Content:    result.Content,
			CreateTime: result.CreateTime,
		})
	}

	return commentList, nil
}

// ListMoment 获取用户的朋友圈动态列表
// 基于 timeline 表查询，按时间倒序展示
func ListMoment(userID int) ([]*dto.MomentListResp, error) {
	// 查询用户的 timeline 记录，关联 moment 和 users 表
	var results []struct {
		MomentID   int64     `gorm:"column:moment_id"`
		UserID     int64     `gorm:"column:user_id"`
		Username   string    `gorm:"column:username"`
		Content    string    `gorm:"column:content"`
		CreateTime time.Time `gorm:"column:created_at"`
	}

	// 从 timeline 表查询用户的动态，按时间倒序
	if err := dao.DB.Table("timeline t").
		Select("m.id as moment_id, m.user_id, u.username, m.content, m.created_at").
		Joins("JOIN moment m ON t.moment_id = m.id").
		Joins("JOIN users u ON m.user_id = u.id").
		Where("t.user_id = ? AND m.delete_time IS NULL", userID).
		Order("t.created_at DESC").
		Scan(&results).Error; err != nil {
		return nil, err
	}

	// 转换为响应格式
	var momentList []*dto.MomentListResp
	for _, result := range results {
		// 根据动态ID查询评论列表
		commentList, err := GetCommentsByMomentID(result.MomentID)
		if err != nil {
			return nil, err
		}
		// 查询动态的点赞数量
		var likeCount int64
		if err := dao.DB.Model(&model.Like{}).
			Where("moment_id = ?", result.MomentID).
			Count(&likeCount).Error; err != nil {
			return nil, err
		}

		momentList = append(momentList, &dto.MomentListResp{
			MomentID:    result.MomentID,
			UserID:      result.UserID,
			Username:    result.Username,
			Content:     result.Content,
			LikeCount:   likeCount,
			CommentList: commentList,
			CreateTime:  result.CreateTime,
		})
	}

	return momentList, nil
}
