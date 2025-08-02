package service

import "errors"

// 服务层错误定义
var (
	// 连接相关错误
	ErrConnectionPoolFull    = errors.New("连接池已满，无法接受新连接")
	ErrClientNotFound        = errors.New("客户端未找到")
	ErrClientAlreadyExists   = errors.New("客户端已存在")
	ErrConnectionClosed      = errors.New("连接已关闭")
	ErrConnectionTimeout     = errors.New("连接超时")
	
	// 消息相关错误
	ErrMessageTooLarge       = errors.New("消息过大")
	ErrMessageQueueFull      = errors.New("消息队列已满")
	ErrMessageSendTimeout    = errors.New("消息发送超时")
	ErrInvalidMessageFormat  = errors.New("无效的消息格式")
	ErrMessageSerializeFailed = errors.New("消息序列化失败")
	
	// 分片相关错误
	ErrShardNotFound         = errors.New("分片未找到")
	ErrShardOverloaded       = errors.New("分片过载")
	ErrShardStopped          = errors.New("分片已停止")
	
	// 认证相关错误
	ErrUnauthorized          = errors.New("未授权访问")
	ErrInvalidToken          = errors.New("无效的令牌")
	ErrTokenExpired          = errors.New("令牌已过期")
	
	// 业务逻辑错误
	ErrUserNotOnline         = errors.New("用户不在线")
	ErrGroupNotFound         = errors.New("群组未找到")
	ErrPermissionDenied      = errors.New("权限不足")
	ErrRateLimitExceeded     = errors.New("请求频率超限")
)

// ErrorCode 错误码定义
type ErrorCode int

const (
	// 成功
	CodeSuccess ErrorCode = 0
	
	// 客户端错误 (1000-1999)
	CodeInvalidRequest     ErrorCode = 1000
	CodeInvalidParameter   ErrorCode = 1001
	CodeUnauthorized       ErrorCode = 1002
	CodePermissionDenied   ErrorCode = 1003
	CodeRateLimitExceeded  ErrorCode = 1004
	
	// 服务器错误 (2000-2999)
	CodeInternalError      ErrorCode = 2000
	CodeServiceUnavailable ErrorCode = 2001
	CodeDatabaseError      ErrorCode = 2002
	CodeKafkaError         ErrorCode = 2003
	CodeRedisError         ErrorCode = 2004
	
	// 连接错误 (3000-3999)
	CodeConnectionFailed   ErrorCode = 3000
	CodeConnectionTimeout  ErrorCode = 3001
	CodeConnectionPoolFull ErrorCode = 3002
	CodeClientNotFound     ErrorCode = 3003
	
	// 消息错误 (4000-4999)
	CodeMessageTooLarge    ErrorCode = 4000
	CodeMessageQueueFull   ErrorCode = 4001
	CodeMessageSendFailed  ErrorCode = 4002
	CodeInvalidMessageFormat ErrorCode = 4003
)

// ServiceError 服务错误结构
type ServiceError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

// Error 实现error接口
func (e *ServiceError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

// NewServiceError 创建服务错误
func NewServiceError(code ErrorCode, message string, details ...string) *ServiceError {
	err := &ServiceError{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}

// IsRetryableError 判断是否为可重试错误
func IsRetryableError(err error) bool {
	if serviceErr, ok := err.(*ServiceError); ok {
		switch serviceErr.Code {
		case CodeServiceUnavailable, CodeConnectionTimeout, CodeMessageQueueFull:
			return true
		}
	}
	return false
}

// IsTemporaryError 判断是否为临时错误
func IsTemporaryError(err error) bool {
	if serviceErr, ok := err.(*ServiceError); ok {
		switch serviceErr.Code {
		case CodeConnectionTimeout, CodeMessageQueueFull, CodeServiceUnavailable:
			return true
		}
	}
	return false
}