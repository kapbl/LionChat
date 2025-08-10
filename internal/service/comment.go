package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"errors"
	"time"
)

func CreateComment(userID, momentID int64, content string) error {
	if momentID < 0 {
		return errors.New("评论对象不存在")
	}
	if content == "" {
		return errors.New("评论内容不能为空")
	}
	comment := &model.Comment{
		MomentID: momentID,
		Content:  content,
		UserID:   userID,
		CreateAt: time.Now(),
	}
	if err := dao.DB.Table("comment").Create(comment).Error; err != nil {
		return err
	}
	return nil
}

func LikeComment(userID, momentID int64) error {
	if momentID < 0 {
		return errors.New("评论对象不存在")
	}

	comment := &model.Like{
		MomentID: momentID,
		UserID:   userID,
		CreateAt: time.Now(),
	}
	if err := dao.DB.Table("like").Create(comment).Error; err != nil {
		return err
	}
	return nil
}
