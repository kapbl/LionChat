package dto

import "time"

// ========== 动态创建和管理相关 DTO ==========

// CreateMomentRequest 创建动态请求
type CreateMomentRequest struct {
	BaseRequest
	Content    string   `json:"content" validate:"required,min=1,max=2000" example:"今天天气真好！"`
	Images     []string `json:"images" validate:"omitempty,max=9,dive,url" example:"[\"https://example.com/image1.jpg\"]"`
	Location   string   `json:"location" validate:"omitempty,max=100" example:"北京市朝阳区"`
	Visibility string   `json:"visibility" validate:"omitempty,oneof=public friends private" default:"public"`
	Tags       []string `json:"tags" validate:"omitempty,max=10,dive,max=20" example:"[\"生活\",\"分享\"]"`
}

// UpdateMomentRequest 更新动态请求
type UpdateMomentRequest struct {
	BaseRequest
	MomentID   int64    `json:"moment_id" validate:"required"`
	Content    string   `json:"content" validate:"omitempty,min=1,max=2000"`
	Images     []string `json:"images" validate:"omitempty,max=9,dive,url"`
	Location   string   `json:"location" validate:"omitempty,max=100"`
	Visibility string   `json:"visibility" validate:"omitempty,oneof=public friends private"`
	Tags       []string `json:"tags" validate:"omitempty,max=10,dive,max=20"`
}

// DeleteMomentRequest 删除动态请求
type DeleteMomentRequest struct {
	BaseRequest
	MomentID int64  `json:"moment_id" validate:"required"`
	Reason   string `json:"reason" validate:"omitempty,max=200"`
}

// ========== 动态互动相关 DTO ==========

// LikeMomentRequest 点赞动态请求
type LikeMomentRequest struct {
	BaseRequest
	MomentID int64 `json:"moment_id" validate:"required"`
	Action   string `json:"action" validate:"required,oneof=like unlike" example:"like"`
}

// ShareMomentRequest 分享动态请求
type ShareMomentRequest struct {
	BaseRequest
	MomentID int64  `json:"moment_id" validate:"required"`
	Content  string `json:"content" validate:"omitempty,max=500" example:"分享一个有趣的动态"`
	Target   string `json:"target" validate:"required,oneof=timeline group friend" example:"timeline"`
	TargetID string `json:"target_id" validate:"omitempty" example:"group_uuid_or_friend_uuid"`
}

// ReportMomentRequest 举报动态请求
type ReportMomentRequest struct {
	BaseRequest
	MomentID int64  `json:"moment_id" validate:"required"`
	Reason   string `json:"reason" validate:"required,oneof=spam inappropriate harassment copyright other" example:"spam"`
	Details  string `json:"details" validate:"omitempty,max=500"`
}

// ========== 动态查询相关 DTO ==========

// MomentListRequest 动态列表请求
type MomentListRequest struct {
	BaseRequest
	UserID     *int64 `json:"user_id" validate:"omitempty"`                                    // 指定用户ID，为空则获取时间线
	Type       string `json:"type" validate:"omitempty,oneof=timeline user following" default:"timeline"` // timeline:时间线 user:用户动态 following:关注的人
	Visibility string `json:"visibility" validate:"omitempty,oneof=public friends private"`
	SinceID    *int64 `json:"since_id" validate:"omitempty"`    // 获取指定ID之后的动态
	MaxID      *int64 `json:"max_id" validate:"omitempty"`      // 获取指定ID之前的动态
	Limit      int    `json:"limit" validate:"omitempty,min=1,max=50" default:"20"`
}

// MomentDetailRequest 动态详情请求
type MomentDetailRequest struct {
	BaseRequest
	MomentID int64 `json:"moment_id" validate:"required"`
}

// MomentSearchRequest 搜索动态请求
type MomentSearchRequest struct {
	BaseRequest
	Keyword    string `json:"keyword" validate:"required,min=1,max=100"`
	UserID     *int64 `json:"user_id" validate:"omitempty"`
	Tags       []string `json:"tags" validate:"omitempty,max=5,dive,max=20"`
	StartTime  *time.Time `json:"start_time" validate:"omitempty"`
	EndTime    *time.Time `json:"end_time" validate:"omitempty"`
	Page       int    `json:"page" validate:"omitempty,min=1" default:"1"`
	PageSize   int    `json:"page_size" validate:"omitempty,min=1,max=50" default:"20"`
}

// ========== 响应数据结构 ==========

// MomentProfile 动态详细信息
type MomentProfile struct {
	ID           int64         `json:"id"`
	Author       UserMixin     `json:"author"`
	Content      string        `json:"content"`
	Images       []MomentImage `json:"images,omitempty"`
	Location     string        `json:"location,omitempty"`
	Visibility   string        `json:"visibility"`
	Tags         []string      `json:"tags,omitempty"`
	LikeCount    int64         `json:"like_count"`
	CommentCount int64         `json:"comment_count"`
	ShareCount   int64         `json:"share_count"`
	ViewCount    int64         `json:"view_count"`
	IsLiked      bool          `json:"is_liked"`      // 当前用户是否已点赞
	IsBookmarked bool          `json:"is_bookmarked"` // 当前用户是否已收藏
	CanEdit      bool          `json:"can_edit"`      // 当前用户是否可编辑
	CanDelete    bool          `json:"can_delete"`    // 当前用户是否可删除
	TimestampMixin
}

// MomentSummary 动态摘要信息 (用于列表展示)
type MomentSummary struct {
	ID           int64         `json:"id"`
	Author       UserMixin     `json:"author"`
	Content      string        `json:"content"`
	Images       []MomentImage `json:"images,omitempty"`
	Location     string        `json:"location,omitempty"`
	Tags         []string      `json:"tags,omitempty"`
	LikeCount    int64         `json:"like_count"`
	CommentCount int64         `json:"comment_count"`
	IsLiked      bool          `json:"is_liked"`
	CreatedAt    time.Time     `json:"created_at"`
}

// MomentImage 动态图片信息
type MomentImage struct {
	ID       int64  `json:"id"`
	URL      string `json:"url"`
	ThumbURL string `json:"thumb_url,omitempty"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Format   string `json:"format,omitempty"`
}

// MomentLike 动态点赞信息
type MomentLike struct {
	ID        int64     `json:"id"`
	User      UserMixin `json:"user"`
	MomentID  int64     `json:"moment_id"`
	CreatedAt time.Time `json:"created_at"`
}

// MomentShare 动态分享信息
type MomentShare struct {
	ID        int64     `json:"id"`
	User      UserMixin `json:"user"`
	MomentID  int64     `json:"moment_id"`
	Content   string    `json:"content,omitempty"`
	Target    string    `json:"target"`
	TargetID  string    `json:"target_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ========== 统计相关 DTO ==========

// MomentStats 动态统计信息
type MomentStats struct {
	TotalMoments   int64 `json:"total_moments"`
	TotalLikes     int64 `json:"total_likes"`
	TotalComments  int64 `json:"total_comments"`
	TotalShares    int64 `json:"total_shares"`
	TotalViews     int64 `json:"total_views"`
	PopularMoments int64 `json:"popular_moments"`
	TrendingTags   []string `json:"trending_tags"`
}
