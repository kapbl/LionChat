package dto

import (
	"time"
)

// APIResponse 统一的API响应结构 (参考GitHub API、GitLab API设计)
type APIResponse[T any] struct {
	Code      int    `json:"code"`                // 业务状态码
	Message   string `json:"message"`             // 响应消息
	Data      T      `json:"data,omitempty"`      // 响应数据
	Timestamp int64  `json:"timestamp"`           // 时间戳
	RequestID string `json:"request_id,omitempty"` // 请求ID，用于追踪
}

// PageInfo 分页信息 (参考GitHub API分页设计)
type PageInfo struct {
	Page     int `json:"page"`      // 当前页码
	PageSize int `json:"page_size"` // 每页大小
	Total    int `json:"total"`     // 总记录数
	Pages    int `json:"pages"`     // 总页数
}

// PagedResponse 分页响应结构
type PagedResponse[T any] struct {
	Code      int      `json:"code"`
	Message   string   `json:"message"`
	Data      []T      `json:"data"`
	PageInfo  PageInfo `json:"page_info"`
	Timestamp int64    `json:"timestamp"`
	RequestID string   `json:"request_id,omitempty"`
}

// ErrorDetail 错误详情 (参考RFC 7807 Problem Details)
type ErrorDetail struct {
	Field   string `json:"field,omitempty"`   // 错误字段
	Message string `json:"message"`           // 错误消息
	Code    string `json:"code,omitempty"`    // 错误代码
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Code      int           `json:"code"`
	Message   string        `json:"message"`
	Errors    []ErrorDetail `json:"errors,omitempty"`
	Timestamp int64         `json:"timestamp"`
	RequestID string        `json:"request_id,omitempty"`
}

// BaseRequest 基础请求结构
type BaseRequest struct {
	RequestID string `json:"request_id,omitempty" validate:"-"` // 请求ID
}

// TimestampMixin 时间戳混入
type TimestampMixin struct {
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// UserMixin 用户信息混入
type UserMixin struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}
