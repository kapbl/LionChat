package dto

import "time"

// ========== 好友搜索相关 DTO ==========

// FriendSearchRequest 搜索好友请求
type FriendSearchRequest struct {
	BaseRequest
	Keyword  string `json:"keyword" validate:"required,min=1,max=100" example:"john"`
	Page     int    `json:"page" validate:"omitempty,min=1" default:"1"`
	PageSize int    `json:"page_size" validate:"omitempty,min=1,max=50" default:"20"`
}

// ========== 好友请求相关 DTO ==========

// SendFriendRequestRequest 发送好友请求
type SendFriendRequestRequest struct {
	BaseRequest
	TargetUserID string `json:"target_user_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Message      string `json:"message" validate:"omitempty,max=200" example:"你好，我想加你为好友"`
	Source       string `json:"source" validate:"omitempty,oneof=search qr_code phone contacts" default:"search"`
}

// HandleFriendRequestRequest 处理好友请求
type HandleFriendRequestRequest struct {
	BaseRequest
	RequestID int64  `json:"request_id" validate:"required"`
	Action    string `json:"action" validate:"required,oneof=accept reject" example:"accept"`
	Message   string `json:"message" validate:"omitempty,max=200" example:"很高兴认识你"`
}

// ========== 好友管理相关 DTO ==========

// UpdateFriendRequest 更新好友信息请求
type UpdateFriendRequest struct {
	BaseRequest
	FriendID string `json:"friend_id" validate:"required"`
	Alias    string `json:"alias" validate:"omitempty,max=50" example:"小明"`
	Tags     string `json:"tags" validate:"omitempty,max=200" example:"同事,朋友"`
	Memo     string `json:"memo" validate:"omitempty,max=500" example:"公司同事，技术很好"`
}

// DeleteFriendRequest 删除好友请求
type DeleteFriendRequest struct {
	BaseRequest
	FriendID string `json:"friend_id" validate:"required"`
	Reason   string `json:"reason" validate:"omitempty,max=200"`
}

// BlockFriendRequest 拉黑好友请求
type BlockFriendRequest struct {
	BaseRequest
	FriendID string `json:"friend_id" validate:"required"`
	Reason   string `json:"reason" validate:"omitempty,max=200"`
}

// ========== 好友列表相关 DTO ==========

// FriendListRequest 好友列表请求
type FriendListRequest struct {
	BaseRequest
	Status   *int   `json:"status" validate:"omitempty,oneof=0 1 2 3"` // 0:正常 1:已删除 2:已拉黑 3:待确认
	Keyword  string `json:"keyword" validate:"omitempty,max=100"`
	Page     int    `json:"page" validate:"omitempty,min=1" default:"1"`
	PageSize int    `json:"page_size" validate:"omitempty,min=1,max=100" default:"50"`
}

// FriendRequestListRequest 好友请求列表请求
type FriendRequestListRequest struct {
	BaseRequest
	Type     string `json:"type" validate:"omitempty,oneof=sent received" default:"received"` // sent:发送的 received:接收的
	Status   *int   `json:"status" validate:"omitempty,oneof=0 1 2"`                         // 0:待处理 1:已同意 2:已拒绝
	Page     int    `json:"page" validate:"omitempty,min=1" default:"1"`
	PageSize int    `json:"page_size" validate:"omitempty,min=1,max=50" default:"20"`
}

// ========== 响应数据结构 ==========

// FriendProfile 好友档案信息
type FriendProfile struct {
	UserMixin
	Alias       string     `json:"alias,omitempty"`        // 好友备注
	Tags        string     `json:"tags,omitempty"`         // 好友标签
	Memo        string     `json:"memo,omitempty"`         // 好友备忘
	Status      int        `json:"status"`                 // 0:正常 1:已删除 2:已拉黑
	IsMutual    bool       `json:"is_mutual"`              // 是否互为好友
	Source      string     `json:"source,omitempty"`       // 添加来源
	AddedAt     time.Time  `json:"added_at"`               // 添加时间
	LastChatAt  *time.Time `json:"last_chat_at,omitempty"` // 最后聊天时间
	ChatCount   int64      `json:"chat_count"`             // 聊天次数
	IsOnline    bool       `json:"is_online"`              // 是否在线
	LastSeenAt  *time.Time `json:"last_seen_at,omitempty"` // 最后在线时间
	TimestampMixin
}

// FriendSummary 好友摘要信息 (用于列表展示)
type FriendSummary struct {
	UserMixin
	Alias      string     `json:"alias,omitempty"`
	Status     int        `json:"status"`
	IsOnline   bool       `json:"is_online"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
}

// FriendRequest 好友请求信息
type FriendRequest struct {
	ID           int64     `json:"id"`
	FromUser     UserMixin `json:"from_user"`
	ToUser       UserMixin `json:"to_user"`
	Message      string    `json:"message,omitempty"`
	Source       string    `json:"source,omitempty"`
	Status       int       `json:"status"`        // 0:待处理 1:已同意 2:已拒绝
	HandledAt    *time.Time `json:"handled_at,omitempty"`
	HandleMessage string   `json:"handle_message,omitempty"`
	TimestampMixin
}

// FriendRequestSummary 好友请求摘要
type FriendRequestSummary struct {
	ID        int64     `json:"id"`
	User      UserMixin `json:"user"` // 根据请求类型，可能是发送者或接收者
	Message   string    `json:"message,omitempty"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// ========== 统计相关 DTO ==========

// FriendStats 好友统计信息
type FriendStats struct {
	TotalFriends    int64 `json:"total_friends"`
	OnlineFriends   int64 `json:"online_friends"`
	PendingRequests int64 `json:"pending_requests"`
	BlockedFriends  int64 `json:"blocked_friends"`
	MutualFriends   int64 `json:"mutual_friends"`
}
