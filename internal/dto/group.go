package dto

import "time"

// ========== 群组创建和管理相关 DTO ==========

// CreateGroupRequest 创建群组请求
type CreateGroupRequest struct {
	BaseRequest
	Name        string `json:"name" validate:"required,min=1,max=50" example:"技术交流群"`
	Description string `json:"description" validate:"omitempty,max=500" example:"技术讨论和分享"`
	Type        string `json:"type" validate:"required,oneof=public private" example:"public"`
	MaxMembers  int    `json:"max_members" validate:"omitempty,min=2,max=2000" default:"500"`
	Avatar      string `json:"avatar" validate:"omitempty,url"`
	Tags        string `json:"tags" validate:"omitempty,max=200" example:"技术,编程,交流"`
}

// UpdateGroupRequest 更新群组信息请求
type UpdateGroupRequest struct {
	BaseRequest
	GroupID     string `json:"group_id" validate:"required"`
	Name        string `json:"name" validate:"omitempty,min=1,max=50"`
	Description string `json:"description" validate:"omitempty,max=500"`
	Avatar      string `json:"avatar" validate:"omitempty,url"`
	Tags        string `json:"tags" validate:"omitempty,max=200"`
	MaxMembers  *int   `json:"max_members" validate:"omitempty,min=2,max=2000"`
}

// DeleteGroupRequest 删除群组请求
type DeleteGroupRequest struct {
	BaseRequest
	GroupID string `json:"group_id" validate:"required"`
	Reason  string `json:"reason" validate:"omitempty,max=200"`
}

// ========== 群组成员管理相关 DTO ==========

// JoinGroupRequest 加入群组请求
type JoinGroupRequest struct {
	BaseRequest
	GroupID string `json:"group_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Message string `json:"message" validate:"omitempty,max=200" example:"申请加入群组"`
	Source  string `json:"source" validate:"omitempty,oneof=search invite qr_code" default:"search"`
}

// LeaveGroupRequest 离开群组请求
type LeaveGroupRequest struct {
	BaseRequest
	GroupID string `json:"group_id" validate:"required"`
	Reason  string `json:"reason" validate:"omitempty,max=200"`
}

// InviteMemberRequest 邀请成员请求
type InviteMemberRequest struct {
	BaseRequest
	GroupID   string   `json:"group_id" validate:"required"`
	UserIDs   []string `json:"user_ids" validate:"required,min=1,max=50,dive,required"`
	Message   string   `json:"message" validate:"omitempty,max=200"`
	ExpiresAt *time.Time `json:"expires_at" validate:"omitempty"`
}

// RemoveMemberRequest 移除成员请求
type RemoveMemberRequest struct {
	BaseRequest
	GroupID string `json:"group_id" validate:"required"`
	UserID  string `json:"user_id" validate:"required"`
	Reason  string `json:"reason" validate:"omitempty,max=200"`
}

// UpdateMemberRoleRequest 更新成员角色请求
type UpdateMemberRoleRequest struct {
	BaseRequest
	GroupID string `json:"group_id" validate:"required"`
	UserID  string `json:"user_id" validate:"required"`
	Role    string `json:"role" validate:"required,oneof=member admin owner" example:"admin"`
}

// ========== 群组搜索和列表相关 DTO ==========

// GroupSearchRequest 搜索群组请求
type GroupSearchRequest struct {
	BaseRequest
	Keyword  string `json:"keyword" validate:"required,min=1,max=100"`
	Type     string `json:"type" validate:"omitempty,oneof=public private"`
	Page     int    `json:"page" validate:"omitempty,min=1" default:"1"`
	PageSize int    `json:"page_size" validate:"omitempty,min=1,max=50" default:"20"`
}

// MyGroupsRequest 我的群组列表请求
type MyGroupsRequest struct {
	BaseRequest
	Role     string `json:"role" validate:"omitempty,oneof=member admin owner"`
	Status   *int   `json:"status" validate:"omitempty,oneof=0 1 2"` // 0:正常 1:已退出 2:已解散
	Page     int    `json:"page" validate:"omitempty,min=1" default:"1"`
	PageSize int    `json:"page_size" validate:"omitempty,min=1,max=100" default:"50"`
}

// GroupMembersRequest 群组成员列表请求
type GroupMembersRequest struct {
	BaseRequest
	GroupID  string `json:"group_id" validate:"required"`
	Role     string `json:"role" validate:"omitempty,oneof=member admin owner"`
	Keyword  string `json:"keyword" validate:"omitempty,max=100"`
	Page     int    `json:"page" validate:"omitempty,min=1" default:"1"`
	PageSize int    `json:"page_size" validate:"omitempty,min=1,max=100" default:"50"`
}

// ========== 响应数据结构 ==========

// GroupProfile 群组档案信息
type GroupProfile struct {
	ID          int64      `json:"id"`
	UUID        string     `json:"uuid"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	Avatar      string     `json:"avatar,omitempty"`
	Type        string     `json:"type"`        // public, private
	Status      int        `json:"status"`      // 0:正常 1:禁用 2:解散
	MaxMembers  int        `json:"max_members"`
	MemberCount int        `json:"member_count"`
	Owner       UserMixin  `json:"owner"`
	Tags        string     `json:"tags,omitempty"`
	Settings    GroupSettings `json:"settings"`
	MyRole      string     `json:"my_role,omitempty"` // 当前用户在群组中的角色
	JoinedAt    *time.Time `json:"joined_at,omitempty"`
	TimestampMixin
}

// GroupSummary 群组摘要信息 (用于列表展示)
type GroupSummary struct {
	ID          int64     `json:"id"`
	UUID        string    `json:"uuid"`
	Name        string    `json:"name"`
	Avatar      string    `json:"avatar,omitempty"`
	Type        string    `json:"type"`
	MemberCount int       `json:"member_count"`
	MyRole      string    `json:"my_role,omitempty"`
	LastMessage *GroupLastMessage `json:"last_message,omitempty"`
	UnreadCount int       `json:"unread_count"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GroupMember 群组成员信息
type GroupMember struct {
	UserMixin
	Role        string     `json:"role"`        // member, admin, owner
	Alias       string     `json:"alias,omitempty"`
	JoinedAt    time.Time  `json:"joined_at"`
	LastSeenAt  *time.Time `json:"last_seen_at,omitempty"`
	IsOnline    bool       `json:"is_online"`
	MessageCount int64     `json:"message_count"`
	IsMuted     bool       `json:"is_muted"`
	MutedUntil  *time.Time `json:"muted_until,omitempty"`
}

// GroupSettings 群组设置
type GroupSettings struct {
	AllowMemberInvite   bool `json:"allow_member_invite"`   // 允许成员邀请
	AllowMemberAtAll    bool `json:"allow_member_at_all"`   // 允许成员@所有人
	RequireApproval     bool `json:"require_approval"`      // 需要审批加入
	AllowAnonymous      bool `json:"allow_anonymous"`       // 允许匿名消息
	MessageHistoryDays  int  `json:"message_history_days"`  // 消息历史保留天数
	AutoDeleteInactive  bool `json:"auto_delete_inactive"`  // 自动删除不活跃成员
	InactiveDays        int  `json:"inactive_days"`         // 不活跃天数阈值
}

// GroupLastMessage 群组最后消息
type GroupLastMessage struct {
	ID        int64     `json:"id"`
	Sender    UserMixin `json:"sender"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

// GroupInvitation 群组邀请信息
type GroupInvitation struct {
	ID        int64      `json:"id"`
	Group     GroupSummary `json:"group"`
	Inviter   UserMixin  `json:"inviter"`
	Invitee   UserMixin  `json:"invitee"`
	Message   string     `json:"message,omitempty"`
	Status    int        `json:"status"`    // 0:待处理 1:已接受 2:已拒绝 3:已过期
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	TimestampMixin
}

// ========== 统计相关 DTO ==========

// GroupStats 群组统计信息
type GroupStats struct {
	TotalGroups     int64 `json:"total_groups"`
	OwnedGroups     int64 `json:"owned_groups"`
	AdminGroups     int64 `json:"admin_groups"`
	ActiveGroups    int64 `json:"active_groups"`
	TotalMembers    int64 `json:"total_members"`
	TotalMessages   int64 `json:"total_messages"`
	PendingInvites  int64 `json:"pending_invites"`
}
