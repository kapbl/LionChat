package dto

import "time"

// ========== 评论创建和管理相关 DTO ==========

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	BaseRequest
	MomentID     int64  `json:"moment_id" validate:"required"`
	Content      string `json:"content" validate:"required,min=1,max=1000" example:"很棒的分享！"`
	ParentID     *int64 `json:"parent_id" validate:"omitempty"` // 父评论ID，用于回复评论
	ReplyToUser  *int64 `json:"reply_to_user" validate:"omitempty"` // 回复的用户ID
	Images       []string `json:"images" validate:"omitempty,max=3,dive,url"`
	MentionUsers []int64 `json:"mention_users" validate:"omitempty,max=10,dive,required"`
}

// UpdateCommentRequest 更新评论请求
type UpdateCommentRequest struct {
	BaseRequest
	CommentID int64  `json:"comment_id" validate:"required"`
	Content   string `json:"content" validate:"required,min=1,max=1000"`
	Images    []string `json:"images" validate:"omitempty,max=3,dive,url"`
}

// DeleteCommentRequest 删除评论请求
type DeleteCommentRequest struct {
	BaseRequest
	CommentID int64  `json:"comment_id" validate:"required"`
	Reason    string `json:"reason" validate:"omitempty,max=200"`
}

// ========== 评论互动相关 DTO ==========

// LikeCommentRequest 点赞评论请求
type LikeCommentRequest struct {
	BaseRequest
	CommentID int64  `json:"comment_id" validate:"required"`
	Action    string `json:"action" validate:"required,oneof=like unlike" example:"like"`
}

// ReportCommentRequest 举报评论请求
type ReportCommentRequest struct {
	BaseRequest
	CommentID int64  `json:"comment_id" validate:"required"`
	Reason    string `json:"reason" validate:"required,oneof=spam inappropriate harassment other" example:"spam"`
	Details   string `json:"details" validate:"omitempty,max=500"`
}

// ========== 评论查询相关 DTO ==========

// CommentListRequest 评论列表请求
type CommentListRequest struct {
	BaseRequest
	MomentID  int64  `json:"moment_id" validate:"required"`
	ParentID  *int64 `json:"parent_id" validate:"omitempty"` // 获取指定父评论的子评论
	SortBy    string `json:"sort_by" validate:"omitempty,oneof=time likes" default:"time"` // time:按时间排序 likes:按点赞数排序
	Order     string `json:"order" validate:"omitempty,oneof=asc desc" default:"asc"`
	SinceID   *int64 `json:"since_id" validate:"omitempty"`
	MaxID     *int64 `json:"max_id" validate:"omitempty"`
	Limit     int    `json:"limit" validate:"omitempty,min=1,max=100" default:"20"`
}

// CommentDetailRequest 评论详情请求
type CommentDetailRequest struct {
	BaseRequest
	CommentID int64 `json:"comment_id" validate:"required"`
}

// UserCommentsRequest 用户评论列表请求
type UserCommentsRequest struct {
	BaseRequest
	UserID   int64 `json:"user_id" validate:"required"`
	Page     int   `json:"page" validate:"omitempty,min=1" default:"1"`
	PageSize int   `json:"page_size" validate:"omitempty,min=1,max=50" default:"20"`
}

// ========== 响应数据结构 ==========

// CommentProfile 评论详细信息
type CommentProfile struct {
	ID           int64           `json:"id"`
	MomentID     int64           `json:"moment_id"`
	Author       UserMixin       `json:"author"`
	Content      string          `json:"content"`
	Images       []CommentImage  `json:"images,omitempty"`
	ParentID     *int64          `json:"parent_id,omitempty"`
	ReplyToUser  *UserMixin      `json:"reply_to_user,omitempty"`
	MentionUsers []UserMixin     `json:"mention_users,omitempty"`
	LikeCount    int64           `json:"like_count"`
	ReplyCount   int64           `json:"reply_count"`
	IsLiked      bool            `json:"is_liked"`
	CanEdit      bool            `json:"can_edit"`
	CanDelete    bool            `json:"can_delete"`
	Replies      []CommentSummary `json:"replies,omitempty"` // 子评论列表（通常只显示前几条）
	TimestampMixin
}

// CommentSummary 评论摘要信息 (用于列表展示)
type CommentSummary struct {
	ID          int64      `json:"id"`
	MomentID    int64      `json:"moment_id"`
	Author      UserMixin  `json:"author"`
	Content     string     `json:"content"`
	Images      []CommentImage `json:"images,omitempty"`
	ParentID    *int64     `json:"parent_id,omitempty"`
	ReplyToUser *UserMixin `json:"reply_to_user,omitempty"`
	LikeCount   int64      `json:"like_count"`
	ReplyCount  int64      `json:"reply_count"`
	IsLiked     bool       `json:"is_liked"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CommentImage 评论图片信息
type CommentImage struct {
	ID       int64  `json:"id"`
	URL      string `json:"url"`
	ThumbURL string `json:"thumb_url,omitempty"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Format   string `json:"format,omitempty"`
}

// CommentLike 评论点赞信息
type CommentLike struct {
	ID        int64     `json:"id"`
	User      UserMixin `json:"user"`
	CommentID int64     `json:"comment_id"`
	CreatedAt time.Time `json:"created_at"`
}

// CommentTree 评论树结构 (用于展示评论及其回复的层级关系)
type CommentTree struct {
	Comment  CommentProfile `json:"comment"`
	Children []CommentTree  `json:"children,omitempty"`
	Level    int            `json:"level"` // 评论层级，0为顶级评论
}

// ========== 统计相关 DTO ==========

// CommentStats 评论统计信息
type CommentStats struct {
	TotalComments    int64 `json:"total_comments"`
	TotalLikes       int64 `json:"total_likes"`
	TotalReplies     int64 `json:"total_replies"`
	ActiveCommenters int64 `json:"active_commenters"`
	AverageLength    float64 `json:"average_length"`
	PopularComments  int64 `json:"popular_comments"`
}
