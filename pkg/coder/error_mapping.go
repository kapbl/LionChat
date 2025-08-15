package coder

// ErrorMessage 错误码对应的错误消息映射
var ErrorMessage = map[int]string{
	// ========== 通用错误 (1000-1099) ==========
	JSON_BIND_ERROR:         "JSON绑定错误",
	INVALID_PARAMETER:       "参数无效",
	REQUEST_PARAMETER_ERROR: "请求参数错误",
	INVALID_REQUEST_FORMAT:  "请求格式无效",
	MISSING_REQUIRED_FIELD:  "缺少必填字段",
	INVALID_DATA_TYPE:       "数据类型无效",
	DATA_VALIDATION_FAILED:  "数据验证失败",
	OPERATION_NOT_ALLOWED:   "操作不被允许",
	RESOURCE_NOT_FOUND:      "资源未找到",
	DUPLICATE_OPERATION:     "重复操作",

	// ========== 用户相关错误 (1100-1199) ==========
	USER_NOT_FOUND:              "用户不存在",
	USER_ALREADY_EXISTS:         "用户已存在",
	USER_NOT_ONLINE:             "用户不在线",
	USER_ACCOUNT_DISABLED:       "用户账户已禁用",
	USER_ACCOUNT_LOCKED:         "用户账户已锁定",
	USER_PROFILE_UPDATE_FAILED:  "用户资料更新失败",
	USER_AVATAR_UPLOAD_FAILED:   "用户头像上传失败",
	USER_PASSWORD_WEAK:          "用户密码强度不够",
	USER_EMAIL_NOT_VERIFIED:     "用户邮箱未验证",
	USER_PHONE_NOT_VERIFIED:     "用户手机号未验证",
	USER_INFO_INCOMPLETE:        "用户信息不完整",
	USER_PERMISSION_DENIED:      "用户权限不足",
	USER_OPERATION_TOO_FREQUENT: "用户操作过于频繁",
	USER_UUID_EMPTY:             "用户UUID为空",
	USER_USERNAME_INVALID:       "用户名格式无效",
	USER_EMAIL_INVALID:          "邮箱格式无效",

	// ========== 好友相关错误 (1200-1299) ==========
	USER_ALREADY_FRIEND:                "用户已经是好友",
	USER_NOT_FRIEND:                    "用户不是好友",
	USER_FRIEND_REQUEST_ALREADY_SEND:   "用户好友请求已经发送",
	USER_FRIEND_REQUEST_NOT_FOUND:      "用户好友请求不存在",
	USER_FRIEND_REQUEST_ALREADY_ACCEPT: "用户好友请求已经接受",
	USER_FRIEND_REQUEST_ALREADY_REJECT: "用户好友请求已经拒绝",
	USER_FRIEND_REQUEST_ALREADY_CANCEL: "用户好友请求已经取消",
	USER_FRIEND_REQUEST_ALREADY_DELETE: "用户好友请求已经删除",
	FRIEND_REQUEST_EXPIRED:             "好友请求已过期",
	FRIEND_LIST_FULL:                   "好友列表已满",
	CANNOT_ADD_SELF_AS_FRIEND:          "不能添加自己为好友",
	FRIEND_REQUEST_LIMIT_EXCEEDED:      "好友请求次数超限",
	FRIEND_BLACKLISTED:                 "用户已被拉黑",
	FRIEND_SEARCH_FAILED:               "好友搜索失败",

	// ========== 群组相关错误 (1300-1399) ==========
	GROUP_NOT_FOUND:             "群组不存在",
	GROUP_ALREADY_EXISTS:        "群组已存在",
	GROUP_NAME_EMPTY:            "群组名称为空",
	GROUP_NAME_TOO_LONG:         "群组名称过长",
	GROUP_DESCRIPTION_TOO_LONG:  "群组描述过长",
	GROUP_MEMBER_LIMIT_EXCEEDED: "群组成员数量超限",
	GROUP_PERMISSION_DENIED:     "群组权限不足",
	GROUP_MEMBER_NOT_FOUND:      "群组成员不存在",
	GROUP_MEMBER_ALREADY_EXISTS: "群组成员已存在",
	GROUP_OWNER_CANNOT_LEAVE:    "群主不能退出群组",
	GROUP_CREATE_FAILED:         "群组创建失败",
	GROUP_JOIN_FAILED:           "加入群组失败",
	GROUP_LEAVE_FAILED:          "退出群组失败",
	GROUP_DISBANDED:             "群组已解散",
	GROUP_MUTED:                 "群组已被禁言",
	GROUP_MEMBER_MUTED:          "群组成员已被禁言",

	// ========== 消息相关错误 (1400-1499) ==========
	MESSAGE_CONTENT_EMPTY:      "消息内容为空",
	MESSAGE_CONTENT_TOO_LONG:   "消息内容过长",
	MESSAGE_TYPE_INVALID:       "消息类型无效",
	MESSAGE_SEND_FAILED:        "消息发送失败",
	MESSAGE_NOT_FOUND:          "消息不存在",
	MESSAGE_ALREADY_READ:       "消息已读",
	MESSAGE_RECALL_FAILED:      "消息撤回失败",
	MESSAGE_RECALL_TIMEOUT:     "消息撤回超时",
	MESSAGE_FORMAT_INVALID:     "消息格式无效",
	MESSAGE_TOO_LARGE:          "消息过大",
	MESSAGE_QUEUE_FULL:         "消息队列已满",
	MESSAGE_SEND_TIMEOUT:       "消息发送超时",
	MESSAGE_SERIALIZE_FAILED:   "消息序列化失败",
	MESSAGE_DESERIALIZE_FAILED: "消息反序列化失败",
	MESSAGE_FRAGMENT_ERROR:     "消息分片错误",
	MESSAGE_CHECKSUM_MISMATCH:  "消息校验和不匹配",
	MESSAGE_REASSEMBLE_FAILED:  "消息重组失败",

	// ========== 动态(Moment)相关错误 (1500-1599) ==========
	MOMENT_CONTENT_EMPTY:       "动态内容为空",
	MOMENT_CONTENT_TOO_LONG:    "动态内容过长",
	MOMENT_NOT_FOUND:           "动态不存在",
	MOMENT_CREATE_FAILED:       "动态创建失败",
	MOMENT_DELETE_FAILED:       "动态删除失败",
	MOMENT_UPDATE_FAILED:       "动态更新失败",
	MOMENT_PERMISSION_DENIED:   "动态权限不足",
	MOMENT_ALREADY_LIKED:       "动态已点赞",
	MOMENT_NOT_LIKED:           "动态未点赞",
	MOMENT_IMAGE_UPLOAD_FAILED: "动态图片上传失败",
	MOMENT_VIDEO_UPLOAD_FAILED: "动态视频上传失败",

	// ========== 评论相关错误 (1600-1699) ==========
	COMMENT_CONTENT_EMPTY:     "评论内容为空",
	COMMENT_CONTENT_TOO_LONG:  "评论内容过长",
	COMMENT_NOT_FOUND:         "评论不存在",
	COMMENT_CREATE_FAILED:     "评论创建失败",
	COMMENT_DELETE_FAILED:     "评论删除失败",
	COMMENT_UPDATE_FAILED:     "评论更新失败",
	COMMENT_PERMISSION_DENIED: "评论权限不足",
	COMMENT_ALREADY_LIKED:     "评论已点赞",
	COMMENT_NOT_LIKED:         "评论未点赞",
	COMMENT_TARGET_NOT_EXIST:  "评论对象不存在",

	// ========== 文件相关错误 (1700-1799) ==========
	FILE_UPLOAD_FAILED:      "文件上传失败",
	FILE_DOWNLOAD_FAILED:    "文件下载失败",
	FILE_NOT_FOUND:          "文件不存在",
	FILE_TYPE_NOT_SUPPORTED: "文件类型不支持",
	FILE_SIZE_EXCEEDED:      "文件大小超限",
	FILE_NAME_INVALID:       "文件名无效",
	FILE_STORAGE_FULL:       "文件存储空间已满",
	FILE_PERMISSION_DENIED:  "文件权限不足",
	FILE_CORRUPTED:          "文件已损坏",
	FILE_VIRUS_DETECTED:     "文件检测到病毒",

	// ========== 认证相关错误 (1800-1899) ==========
	AUTH_TOKEN_INVALID:             "认证令牌无效",
	AUTH_TOKEN_EXPIRED:             "认证令牌已过期",
	AUTH_TOKEN_MISSING:             "认证令牌缺失",
	AUTH_LOGIN_FAILED:              "登录失败",
	AUTH_PASSWORD_INCORRECT:        "密码错误",
	AUTH_ACCOUNT_NOT_EXIST:         "账户不存在",
	AUTH_ACCOUNT_DISABLED:          "账户已禁用",
	AUTH_PERMISSION_DENIED:         "权限不足",
	AUTH_SESSION_EXPIRED:           "会话已过期",
	AUTH_LOGOUT_FAILED:             "登出失败",
	AUTH_REGISTER_FAILED:           "注册失败",
	AUTH_EMAIL_EXISTS:              "邮箱已存在",
	AUTH_USERNAME_EXISTS:           "用户名已存在",
	AUTH_PASSWORD_ENCRYPT_FAILED:   "密码加密失败",
	AUTH_TOKEN_GENERATE_FAILED:     "令牌生成失败",
	AUTH_VERIFICATION_CODE_INVALID: "验证码无效",
	AUTH_VERIFICATION_CODE_EXPIRED: "验证码已过期",

	// ========== 数据库相关错误 (1900-1999) ==========
	DB_CONNECTION_FAILED:    "数据库连接失败",
	DB_QUERY_FAILED:         "数据库查询失败",
	DB_INSERT_FAILED:        "数据库插入失败",
	DB_UPDATE_FAILED:        "数据库更新失败",
	DB_DELETE_FAILED:        "数据库删除失败",
	DB_TRANSACTION_FAILED:   "数据库事务失败",
	DB_RECORD_NOT_FOUND:     "数据库记录不存在",
	DB_DUPLICATE_KEY:        "数据库主键冲突",
	DB_CONSTRAINT_VIOLATION: "数据库约束违反",
	DB_TIMEOUT:              "数据库操作超时",
	DB_POOL_EXHAUSTED:       "数据库连接池耗尽",
	DB_MIGRATION_FAILED:     "数据库迁移失败",

	// ========== 网络连接相关错误 (2000-2099) ==========
	CONNECTION_FAILED:           "连接失败",
	CONNECTION_TIMEOUT:          "连接超时",
	CONNECTION_CLOSED:           "连接已关闭",
	CONNECTION_POOL_FULL:        "连接池已满",
	CLIENT_NOT_FOUND:            "客户端未找到",
	CLIENT_ALREADY_EXISTS:       "客户端已存在",
	WEBSOCKET_CONNECTION_FAILED: "WebSocket连接失败",
	WEBSOCKET_UPGRADE_FAILED:    "WebSocket升级失败",
	NETWORK_UNREACHABLE:         "网络不可达",
	SERVER_UNAVAILABLE:          "服务器不可用",

	// ========== 消息队列相关错误 (2100-2199) ==========
	KAFKA_CONNECTION_FAILED:      "Kafka连接失败",
	KAFKA_PRODUCER_FAILED:        "Kafka生产者失败",
	KAFKA_CONSUMER_FAILED:        "Kafka消费者失败",
	KAFKA_TOPIC_NOT_FOUND:        "Kafka主题不存在",
	KAFKA_MESSAGE_SEND_FAILED:    "Kafka消息发送失败",
	KAFKA_MESSAGE_CONSUME_FAILED: "Kafka消息消费失败",
	KAFKA_PARTITION_ERROR:        "Kafka分区错误",
	KAFKA_OFFSET_ERROR:           "Kafka偏移量错误",

	// ========== 缓存相关错误 (2200-2299) ==========
	REDIS_CONNECTION_FAILED:   "Redis连接失败",
	REDIS_OPERATION_FAILED:    "Redis操作失败",
	REDIS_KEY_NOT_FOUND:       "Redis键不存在",
	REDIS_KEY_EXPIRED:         "Redis键已过期",
	REDIS_MEMORY_FULL:         "Redis内存已满",
	REDIS_CLUSTER_ERROR:       "Redis集群错误",
	CACHE_MISS:                "缓存未命中",
	CACHE_INVALIDATION_FAILED: "缓存失效失败",

	// ========== 系统相关错误 (2300-2399) ==========
	SYSTEM_ERROR:         "系统错误",
	SERVICE_UNAVAILABLE:  "服务不可用",
	RATE_LIMIT_EXCEEDED:  "请求频率超限",
	RESOURCE_EXHAUSTED:   "资源耗尽",
	CONFIG_ERROR:         "配置错误",
	INTERNAL_ERROR:       "内部错误",
	TIMEOUT_ERROR:        "超时错误",
	CONCURRENCY_ERROR:    "并发错误",
	SHARD_NOT_FOUND:      "分片未找到",
	SHARD_OVERLOADED:     "分片过载",
	SHARD_STOPPED:        "分片已停止",
	WORKER_POOL_FULL:     "工作池已满",
	WORKER_NOT_AVAILABLE: "工作者不可用",
	CIRCUIT_BREAKER_OPEN: "熔断器开启",
	HEALTH_CHECK_FAILED:  "健康检查失败",
}

// GetErrorMessage 根据错误码获取错误消息
func GetErrorMessage(code int) string {
	if msg, exists := ErrorMessage[code]; exists {
		return msg
	}
	return "未知错误"
}

// IsRetryableError 判断错误码是否为可重试错误
func IsRetryableError(code int) bool {
	retryableCodes := []int{
		CONNECTION_TIMEOUT,
		CONNECTION_FAILED,
		SERVER_UNAVAILABLE,
		SERVICE_UNAVAILABLE,
		DB_TIMEOUT,
		DB_CONNECTION_FAILED,
		KAFKA_CONNECTION_FAILED,
		KAFKA_MESSAGE_SEND_FAILED,
		REDIS_CONNECTION_FAILED,
		MESSAGE_SEND_TIMEOUT,
		MESSAGE_QUEUE_FULL,
		NETWORK_UNREACHABLE,
		TIMEOUT_ERROR,
	}

	for _, retryableCode := range retryableCodes {
		if code == retryableCode {
			return true
		}
	}
	return false
}

// IsClientError 判断是否为客户端错误（1000-1999）
func IsClientError(code int) bool {
	return code >= 1000 && code < 2000
}

// IsServerError 判断是否为服务端错误（2000-2999）
func IsServerError(code int) bool {
	return code >= 2000 && code < 3000
}

// GetErrorCategory 获取错误类别
func GetErrorCategory(code int) string {
	switch {
	case code >= 1000 && code < 1100:
		return "通用错误"
	case code >= 1100 && code < 1200:
		return "用户相关错误"
	case code >= 1200 && code < 1300:
		return "好友相关错误"
	case code >= 1300 && code < 1400:
		return "群组相关错误"
	case code >= 1400 && code < 1500:
		return "消息相关错误"
	case code >= 1500 && code < 1600:
		return "动态相关错误"
	case code >= 1600 && code < 1700:
		return "评论相关错误"
	case code >= 1700 && code < 1800:
		return "文件相关错误"
	case code >= 1800 && code < 1900:
		return "认证相关错误"
	case code >= 1900 && code < 2000:
		return "数据库相关错误"
	case code >= 2000 && code < 2100:
		return "网络连接相关错误"
	case code >= 2100 && code < 2200:
		return "消息队列相关错误"
	case code >= 2200 && code < 2300:
		return "缓存相关错误"
	case code >= 2300 && code < 2400:
		return "系统相关错误"
	default:
		return "未知类别"
	}
}
