package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/internal/dto"
	"cchat/pkg/cerror"
	"time"

	"gorm.io/gorm"
)

type CommentService struct {
	DB       *gorm.DB
	Username string
	UUID     string
	UserID   int64
}

func NewCommentService(db *gorm.DB, username, uuid string, userID int64) *CommentService {
	return &CommentService{DB: db, Username: username, UUID: uuid, UserID: userID}
}

func (s *CommentService) CreateComment(req *dto.CreateCommentRequest) *cerror.CodeError {

	if req.MomentID < 0 {
		return cerror.NewCodeError(4444, "评论对象不存在")
	}
	if req.Content == "" {
		return cerror.NewCodeError(4444, "评论内容不能为空")
	}
	comment := &model.Comment{
		MomentID:  req.MomentID,
		Content:   req.Content,
		UserID:    s.UserID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: nil,
	}
	if err := dao.DB.Table("comment").Create(comment).Error; err != nil {
		return cerror.NewCodeError(4444, "评论创建失败")
	}
	return nil
}

func (s *CommentService) LikeComment(momentID int64) *cerror.CodeError {
	if momentID < 0 {
		return cerror.NewCodeError(4444, "评论对象不存在")
	}
	comment := &model.Like{
		MomentID:  momentID,
		UserID:    s.UserID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DeletedAt: nil,
	}
	if err := dao.DB.Table("like").Create(comment).Error; err != nil {
		return cerror.NewCodeError(4444, "评论点赞失败")

	}
	return nil
}

func (s *CommentService) GetCommentsList(momentID int64) ([]dto.CommentList, *cerror.CodeError) {
	commentList, err := GetCommentsByMomentID(momentID)
	if err != nil {
		return nil, cerror.NewCodeError(4444, err.Error())
	}
	return commentList, nil
}

// 获取用户的朋友圈动态列表
// 基于 timeline 表查询，按时间倒序展示
// GetCommentsByMomentID 根据动态ID查询评论列表
func GetCommentsByMomentID(momentID int64) ([]dto.CommentList, error) {
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
	var commentList []dto.CommentList
	for _, result := range results {
		commentList = append(commentList, dto.CommentList{
			UserID:     result.UserID,
			Username:   result.Username,
			Content:    result.Content,
			CreateTime: result.CreateTime,
		})
	}

	return commentList, nil
}
