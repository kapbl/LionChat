package dto

type CreateCommentReq struct {
	MomentID int64  `json:"moment_id"` // 动态的ID
	Content  string `json:"content"`   // 评论的内容
}

type LikeCommentReq struct {
	MomentID int64 `json:"moment_id"` // 评论的ID
}

type CommentListReq struct {
	MomentID int64 `json:"moment_id"` // 动态的ID
}
