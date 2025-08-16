package dto

import "time"

type CreateCommentRequest struct {
	MomentID int64  `json:"moment_id"` // 动态的ID
	Content  string `json:"content"`   // 评论的内容
}
type CreateCommentResponse struct {
	BaseResponse
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type LikeCommentRequest struct {
	MomentID int64 `json:"moment_id"` // 评论的ID
}
type LikeCommentResponse struct {
	BaseResponse
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type CommentList struct {
	UserID     int64     `json:"user_id"`
	Username   string    `json:"username"`
	Content    string    `json:"content"`
	CreateTime time.Time `json:"create_time"`
}
type GetCommentList struct {
	BaseResponse
	Code        int           `json:"code"`
	Msg         string        `json:"msg"`
	CommentList []CommentList `json:"comment_list"`
}
