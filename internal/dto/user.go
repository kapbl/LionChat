package dto

import "time"

// ========== 用户认证相关 DTO ==========

// LoginRequest 登录请求 (参考OAuth 2.0规范)
type LoginRequest struct {
	BaseRequest
	Username string `json:"username" validate:"required,min=3,max=50" example:"john_doe"`
	Password string `json:"password" validate:"required,min=6,max=128" example:"password123"`
	ClientID string `json:"client_id,omitempty" validate:"omitempty,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	BaseRequest
	Username string `json:"username" validate:"required,min=3,max=50,alphanum" example:"john_doe"`
	Nickname string `json:"nickname" validate:"required,min=1,max=100" example:"John Doe"`
	Password string `json:"password" validate:"required,min=6,max=128" example:"password123"`
	Email    string `json:"email" validate:"required,email" example:"john@example.com"`
	Phone    string `json:"phone,omitempty" validate:"omitempty,e164" example:"+1234567890"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	BaseRequest
	OldPassword string `json:"old_password" validate:"required,min=6,max=128"`
	NewPassword string `json:"new_password" validate:"required,min=6,max=128"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	BaseRequest
	Email       string `json:"email" validate:"required,email"`
	ResetToken  string `json:"reset_token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6,max=128"`
}

// ========== 用户信息相关 DTO ==========

// UserProfile 用户档案信息
type UserProfile struct {
	ID          int64      `json:"id"`
	UUID        string     `json:"uuid"`
	Username    string     `json:"username"`
	Nickname    string     `json:"nickname"`
	Email       string     `json:"email"`
	Phone       string     `json:"phone,omitempty"`
	Avatar      string     `json:"avatar"`
	Bio         string     `json:"bio,omitempty"`
	Gender      int        `json:"gender"` // 0:未知 1:男 2:女
	Birthday    *time.Time `json:"birthday,omitempty"`
	Location    string     `json:"location,omitempty"`
	Website     string     `json:"website,omitempty"`
	Status      int        `json:"status"` // 0:正常 1:禁用 2:删除
	IsVerified  bool       `json:"is_verified"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	TimestampMixin
}

// UserSummary 用户摘要信息 (用于列表展示)
type UserSummary struct {
	ID         int64  `json:"id"`
	UUID       string `json:"uuid"`
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	IsVerified bool   `json:"is_verified"`
	Status     int    `json:"status"`
}

// UpdateProfileRequest 更新用户档案请求
type UpdateProfileRequest struct {
	BaseRequest
	Nickname string     `json:"nickname" validate:"omitempty,min=1,max=100"`
	Bio      string     `json:"bio" validate:"omitempty,max=500"`
	Gender   *int       `json:"gender" validate:"omitempty,oneof=0 1 2"`
	Birthday *time.Time `json:"birthday" validate:"omitempty"`
	Location string     `json:"location" validate:"omitempty,max=100"`
	Website  string     `json:"website" validate:"omitempty,url"`
}

// ========== 响应数据结构 ==========

// LoginData 登录响应数据
type LoginData struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	TokenType    string      `json:"token_type"` // Bearer
	ExpiresIn    int64       `json:"expires_in"` // 秒
	UserProfile  UserProfile `json:"user_profile"`
}

// TokenRefreshData 刷新令牌响应数据
type TokenRefreshData struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

// ========== 搜索和查询相关 DTO ==========

// UserSearchRequest 用户搜索请求
type UserSearchRequest struct {
	BaseRequest
	Keyword  string `json:"keyword" validate:"required,min=1,max=100"`
	Page     int    `json:"page" validate:"omitempty,min=1" default:"1"`
	PageSize int    `json:"page_size" validate:"omitempty,min=1,max=100" default:"20"`
}

// UserListRequest 用户列表请求
type UserListRequest struct {
	BaseRequest
	Status   *int `json:"status" validate:"omitempty,oneof=0 1 2"`
	Page     int  `json:"page" validate:"omitempty,min=1" default:"1"`
	PageSize int  `json:"page_size" validate:"omitempty,min=1,max=100" default:"20"`
}

// ========== 统计相关 DTO ==========

// UserStats 用户统计信息
type UserStats struct {
	FriendsCount   int64 `json:"friends_count"`
	GroupsCount    int64 `json:"groups_count"`
	MomentsCount   int64 `json:"moments_count"`
	FollowersCount int64 `json:"followers_count"`
	FollowingCount int64 `json:"following_count"`
}
