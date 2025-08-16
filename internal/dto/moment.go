package dto

import "time"

// MomentCreateRequest 新建动态请求
type MomentCreateRequest struct {
	Content string `json:"content"`
}

// MomentCreateResponse 新建动态响应
type MomentCreateResponse struct {
	BaseResponse
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// MomentListResp 动态列表响应
type MomentListResponse struct {
	BaseResponse
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Data []MomentInfo `json:"data"`
}
type MomentInfo struct {
	MomentID    int64             `json:"moment_id"`
	UserID      int64             `json:"user_id"`
	Username    string            `json:"username"`
	Content     string            `json:"content"`
	LikeCount   int64             `json:"like_count"`
	CommentList []CommentListResp `json:"comment_list"`
	CreateTime  time.Time         `json:"create_time"`
}

type CommentListResp struct {
	UserID     int64     `json:"user_id"`
	Username   string    `json:"username"`
	Content    string    `json:"content"`
	CreateTime time.Time `json:"create_time"`
}
