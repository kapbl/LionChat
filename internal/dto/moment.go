package dto

import "time"

// MomentCreateReq 新建动态请求
type MomentCreateReq struct {
	Content string `json:"content"`
}

type MomentCreateResp struct {
	ID int64 `json:"id"`
}

// MomentListResp 动态列表响应
type MomentListResp struct {
	UserID      int64             `json:"user_id"`
	Username    string            `json:"username"`
	Content     string            `json:"content"`
	LikeCount   int64             `json:"like_count"`
	CommentList []CommentListResp `json:"comment_list"`

	CreateTime time.Time `json:"create_time"`
}

type CommentListResp struct {
	UserID     int64     `json:"user_id"`
	Username   string    `json:"username"`
	Content    string    `json:"content"`
	CreateTime time.Time `json:"create_time"`
}
