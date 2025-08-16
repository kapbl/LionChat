package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/internal/dto"
	"cchat/pkg/cerror"
	"cchat/pkg/logger"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MomentService struct {
	UserID   int64
	UserUUID string
	DB       *gorm.DB
}

func NewMomentService(userID int64, userUUID string, db *gorm.DB) *MomentService {
	return &MomentService{
		UserID:   userID,
		UserUUID: userUUID,
		DB:       db,
	}
}
func (s *MomentService) CreateMoment(moment *dto.MomentCreateRequest) *cerror.CodeError {
	// 校验moment
	if moment == nil {
		return cerror.NewCodeError(2222, "moment is nil")
	}
	// 在数据库中插入这条动态
	m := &model.Moment{
		UserID:    s.UserID,
		Content:   moment.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: nil,
	}
	// 插入到自己的动态表中
	if err := s.DB.Create(m).Error; err != nil {
		return cerror.NewCodeError(2222, err.Error())
	}
	// 插入到自己的Timeline表中
	t := &model.Timeline{
		UserID:    s.UserID,
		MomentID:  m.ID,
		IsOwn:     true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: nil,
	}
	if err := s.DB.Create(t).Error; err != nil {
		return cerror.NewCodeError(2222, err.Error())
	}

	// 异步插入所有可见好友的Timeline表
	go func() {
		// 查询用户的所有好友
		var friends []*model.UserFriends
		if err := s.DB.Where("user_id = ?", s.UserID).Find(&friends).Error; err != nil {
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
	return nil
}

// ListMoment 获取用户的朋友圈动态列表
// 基于 timeline 表查询，按时间倒序展示
func (s *MomentService) ListMoment() ([]dto.MomentInfo, *cerror.CodeError) {

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
		Where("t.user_id = ? AND m.delete_time IS NULL", s.UserID).
		Order("t.created_at DESC").
		Scan(&results).Error; err != nil {
		return nil, cerror.NewCodeError(2222, err.Error())
	}

	// 转换为响应格式
	var momentList []dto.MomentInfo
	for _, result := range results {
		// 根据动态ID查询评论列表
		commentList, err := GetCommentsByMomentID(result.MomentID)
		if err != nil {
			return nil, cerror.NewCodeError(2222, err.Error())
		}
		// 查询动态的点赞数量
		var likeCount int64
		if err := dao.DB.Model(&model.Like{}).
			Where("moment_id = ?", result.MomentID).
			Count(&likeCount).Error; err != nil {
			return nil, cerror.NewCodeError(2222, err.Error())
		}

		momentList = append(momentList, dto.MomentInfo{
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
