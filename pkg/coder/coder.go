package coder

// 错误码定义 - 统一的错误码管理
// 错误码范围分配：
// 1000-1099: 通用错误
// 1100-1199: 用户相关错误
// 1200-1299: 好友相关错误
// 1300-1399: 群组相关错误
// 1400-1499: 消息相关错误
// 1500-1599: 动态(Moment)相关错误
// 1600-1699: 评论相关错误
// 1700-1799: 文件相关错误
// 1800-1899: 认证相关错误
// 1900-1999: 数据库相关错误
// 2000-2099: 网络连接相关错误
// 2100-2199: 消息队列相关错误
// 2200-2299: 缓存相关错误
// 2300-2399: 系统相关错误

var (
	// ========== 通用错误 (1000-1099) ==========
	JSON_BIND_ERROR           = 1000 // JSON绑定错误
	INVALID_PARAMETER         = 1001 // 参数无效
	REQUEST_PARAMETER_ERROR   = 1002 // 请求参数错误
	INVALID_REQUEST_FORMAT    = 1003 // 请求格式无效
	MISSING_REQUIRED_FIELD    = 1004 // 缺少必填字段
	INVALID_DATA_TYPE         = 1005 // 数据类型无效
	DATA_VALIDATION_FAILED    = 1006 // 数据验证失败
	OPERATION_NOT_ALLOWED     = 1007 // 操作不被允许
	RESOURCE_NOT_FOUND        = 1008 // 资源未找到
	DUPLICATE_OPERATION       = 1009 // 重复操作

	// ========== 用户相关错误 (1100-1199) ==========
	USER_NOT_FOUND            = 1100 // 用户不存在
	USER_ALREADY_EXISTS       = 1101 // 用户已存在
	USER_NOT_ONLINE           = 1102 // 用户不在线
	USER_ACCOUNT_DISABLED     = 1103 // 用户账户已禁用
	USER_ACCOUNT_LOCKED       = 1104 // 用户账户已锁定
	USER_PROFILE_UPDATE_FAILED = 1105 // 用户资料更新失败
	USER_AVATAR_UPLOAD_FAILED = 1106 // 用户头像上传失败
	USER_PASSWORD_WEAK        = 1107 // 用户密码强度不够
	USER_EMAIL_NOT_VERIFIED   = 1108 // 用户邮箱未验证
	USER_PHONE_NOT_VERIFIED   = 1109 // 用户手机号未验证
	USER_INFO_INCOMPLETE      = 1110 // 用户信息不完整
	USER_PERMISSION_DENIED    = 1111 // 用户权限不足
	USER_OPERATION_TOO_FREQUENT = 1112 // 用户操作过于频繁
	USER_UUID_EMPTY           = 1113 // 用户UUID为空
	USER_USERNAME_INVALID     = 1114 // 用户名格式无效
	USER_EMAIL_INVALID        = 1115 // 邮箱格式无效

	// ========== 好友相关错误 (1200-1299) ==========
	USER_ALREADY_FRIEND                = 1200 // 用户已经是好友
	USER_NOT_FRIEND                    = 1201 // 用户不是好友
	USER_FRIEND_REQUEST_ALREADY_SEND   = 1202 // 用户好友请求已经发送
	USER_FRIEND_REQUEST_NOT_FOUND      = 1203 // 用户好友请求不存在
	USER_FRIEND_REQUEST_ALREADY_ACCEPT = 1204 // 用户好友请求已经接受
	USER_FRIEND_REQUEST_ALREADY_REJECT = 1205 // 用户好友请求已经拒绝
	USER_FRIEND_REQUEST_ALREADY_CANCEL = 1206 // 用户好友请求已经取消
	USER_FRIEND_REQUEST_ALREADY_DELETE = 1207 // 用户好友请求已经删除
	FRIEND_REQUEST_EXPIRED             = 1208 // 好友请求已过期
	FRIEND_LIST_FULL                   = 1209 // 好友列表已满
	CANNOT_ADD_SELF_AS_FRIEND         = 1210 // 不能添加自己为好友
	FRIEND_REQUEST_LIMIT_EXCEEDED      = 1211 // 好友请求次数超限
	FRIEND_BLACKLISTED                 = 1212 // 用户已被拉黑
	FRIEND_SEARCH_FAILED               = 1213 // 好友搜索失败

	// ========== 群组相关错误 (1300-1399) ==========
	GROUP_NOT_FOUND           = 1300 // 群组不存在
	GROUP_ALREADY_EXISTS      = 1301 // 群组已存在
	GROUP_NAME_EMPTY          = 1302 // 群组名称为空
	GROUP_NAME_TOO_LONG       = 1303 // 群组名称过长
	GROUP_DESCRIPTION_TOO_LONG = 1304 // 群组描述过长
	GROUP_MEMBER_LIMIT_EXCEEDED = 1305 // 群组成员数量超限
	GROUP_PERMISSION_DENIED   = 1306 // 群组权限不足
	GROUP_MEMBER_NOT_FOUND    = 1307 // 群组成员不存在
	GROUP_MEMBER_ALREADY_EXISTS = 1308 // 群组成员已存在
	GROUP_OWNER_CANNOT_LEAVE  = 1309 // 群主不能退出群组
	GROUP_CREATE_FAILED       = 1310 // 群组创建失败
	GROUP_JOIN_FAILED         = 1311 // 加入群组失败
	GROUP_LEAVE_FAILED        = 1312 // 退出群组失败
	GROUP_DISBANDED           = 1313 // 群组已解散
	GROUP_MUTED               = 1314 // 群组已被禁言
	GROUP_MEMBER_MUTED        = 1315 // 群组成员已被禁言

	// ========== 消息相关错误 (1400-1499) ==========
	MESSAGE_CONTENT_EMPTY     = 1400 // 消息内容为空
	MESSAGE_CONTENT_TOO_LONG  = 1401 // 消息内容过长
	MESSAGE_TYPE_INVALID      = 1402 // 消息类型无效
	MESSAGE_SEND_FAILED       = 1403 // 消息发送失败
	MESSAGE_NOT_FOUND         = 1404 // 消息不存在
	MESSAGE_ALREADY_READ      = 1405 // 消息已读
	MESSAGE_RECALL_FAILED     = 1406 // 消息撤回失败
	MESSAGE_RECALL_TIMEOUT    = 1407 // 消息撤回超时
	MESSAGE_FORMAT_INVALID    = 1408 // 消息格式无效
	MESSAGE_TOO_LARGE         = 1409 // 消息过大
	MESSAGE_QUEUE_FULL        = 1410 // 消息队列已满
	MESSAGE_SEND_TIMEOUT      = 1411 // 消息发送超时
	MESSAGE_SERIALIZE_FAILED  = 1412 // 消息序列化失败
	MESSAGE_DESERIALIZE_FAILED = 1413 // 消息反序列化失败
	MESSAGE_FRAGMENT_ERROR    = 1414 // 消息分片错误
	MESSAGE_CHECKSUM_MISMATCH = 1415 // 消息校验和不匹配
	MESSAGE_REASSEMBLE_FAILED = 1416 // 消息重组失败

	// ========== 动态(Moment)相关错误 (1500-1599) ==========
	MOMENT_CONTENT_EMPTY      = 1500 // 动态内容为空
	MOMENT_CONTENT_TOO_LONG   = 1501 // 动态内容过长
	MOMENT_NOT_FOUND          = 1502 // 动态不存在
	MOMENT_CREATE_FAILED      = 1503 // 动态创建失败
	MOMENT_DELETE_FAILED      = 1504 // 动态删除失败
	MOMENT_UPDATE_FAILED      = 1505 // 动态更新失败
	MOMENT_PERMISSION_DENIED  = 1506 // 动态权限不足
	MOMENT_ALREADY_LIKED      = 1507 // 动态已点赞
	MOMENT_NOT_LIKED          = 1508 // 动态未点赞
	MOMENT_IMAGE_UPLOAD_FAILED = 1509 // 动态图片上传失败
	MOMENT_VIDEO_UPLOAD_FAILED = 1510 // 动态视频上传失败

	// ========== 评论相关错误 (1600-1699) ==========
	COMMENT_CONTENT_EMPTY     = 1600 // 评论内容为空
	COMMENT_CONTENT_TOO_LONG  = 1601 // 评论内容过长
	COMMENT_NOT_FOUND         = 1602 // 评论不存在
	COMMENT_CREATE_FAILED     = 1603 // 评论创建失败
	COMMENT_DELETE_FAILED     = 1604 // 评论删除失败
	COMMENT_UPDATE_FAILED     = 1605 // 评论更新失败
	COMMENT_PERMISSION_DENIED = 1606 // 评论权限不足
	COMMENT_ALREADY_LIKED     = 1607 // 评论已点赞
	COMMENT_NOT_LIKED         = 1608 // 评论未点赞
	COMMENT_TARGET_NOT_EXIST  = 1609 // 评论对象不存在

	// ========== 文件相关错误 (1700-1799) ==========
	FILE_UPLOAD_FAILED        = 1700 // 文件上传失败
	FILE_DOWNLOAD_FAILED      = 1701 // 文件下载失败
	FILE_NOT_FOUND            = 1702 // 文件不存在
	FILE_TYPE_NOT_SUPPORTED   = 1703 // 文件类型不支持
	FILE_SIZE_EXCEEDED        = 1704 // 文件大小超限
	FILE_NAME_INVALID         = 1705 // 文件名无效
	FILE_STORAGE_FULL         = 1706 // 文件存储空间已满
	FILE_PERMISSION_DENIED    = 1707 // 文件权限不足
	FILE_CORRUPTED            = 1708 // 文件已损坏
	FILE_VIRUS_DETECTED       = 1709 // 文件检测到病毒

	// ========== 认证相关错误 (1800-1899) ==========
	AUTH_TOKEN_INVALID        = 1800 // 认证令牌无效
	AUTH_TOKEN_EXPIRED        = 1801 // 认证令牌已过期
	AUTH_TOKEN_MISSING        = 1802 // 认证令牌缺失
	AUTH_LOGIN_FAILED         = 1803 // 登录失败
	AUTH_PASSWORD_INCORRECT   = 1804 // 密码错误
	AUTH_ACCOUNT_NOT_EXIST    = 1805 // 账户不存在
	AUTH_ACCOUNT_DISABLED     = 1806 // 账户已禁用
	AUTH_PERMISSION_DENIED    = 1807 // 权限不足
	AUTH_SESSION_EXPIRED      = 1808 // 会话已过期
	AUTH_LOGOUT_FAILED        = 1809 // 登出失败
	AUTH_REGISTER_FAILED      = 1810 // 注册失败
	AUTH_EMAIL_EXISTS         = 1811 // 邮箱已存在
	AUTH_USERNAME_EXISTS      = 1812 // 用户名已存在
	AUTH_PASSWORD_ENCRYPT_FAILED = 1813 // 密码加密失败
	AUTH_TOKEN_GENERATE_FAILED = 1814 // 令牌生成失败
	AUTH_VERIFICATION_CODE_INVALID = 1815 // 验证码无效
	AUTH_VERIFICATION_CODE_EXPIRED = 1816 // 验证码已过期

	// ========== 数据库相关错误 (1900-1999) ==========
	DB_CONNECTION_FAILED      = 1900 // 数据库连接失败
	DB_QUERY_FAILED           = 1901 // 数据库查询失败
	DB_INSERT_FAILED          = 1902 // 数据库插入失败
	DB_UPDATE_FAILED          = 1903 // 数据库更新失败
	DB_DELETE_FAILED          = 1904 // 数据库删除失败
	DB_TRANSACTION_FAILED     = 1905 // 数据库事务失败
	DB_RECORD_NOT_FOUND       = 1906 // 数据库记录不存在
	DB_DUPLICATE_KEY          = 1907 // 数据库主键冲突
	DB_CONSTRAINT_VIOLATION   = 1908 // 数据库约束违反
	DB_TIMEOUT                = 1909 // 数据库操作超时
	DB_POOL_EXHAUSTED         = 1910 // 数据库连接池耗尽
	DB_MIGRATION_FAILED       = 1911 // 数据库迁移失败

	// ========== 网络连接相关错误 (2000-2099) ==========
	CONNECTION_FAILED         = 2000 // 连接失败
	CONNECTION_TIMEOUT        = 2001 // 连接超时
	CONNECTION_CLOSED         = 2002 // 连接已关闭
	CONNECTION_POOL_FULL      = 2003 // 连接池已满
	CLIENT_NOT_FOUND          = 2004 // 客户端未找到
	CLIENT_ALREADY_EXISTS     = 2005 // 客户端已存在
	WEBSOCKET_CONNECTION_FAILED = 2006 // WebSocket连接失败
	WEBSOCKET_UPGRADE_FAILED  = 2007 // WebSocket升级失败
	NETWORK_UNREACHABLE       = 2008 // 网络不可达
	SERVER_UNAVAILABLE        = 2009 // 服务器不可用

	// ========== 消息队列相关错误 (2100-2199) ==========
	KAFKA_CONNECTION_FAILED   = 2100 // Kafka连接失败
	KAFKA_PRODUCER_FAILED     = 2101 // Kafka生产者失败
	KAFKA_CONSUMER_FAILED     = 2102 // Kafka消费者失败
	KAFKA_TOPIC_NOT_FOUND     = 2103 // Kafka主题不存在
	KAFKA_MESSAGE_SEND_FAILED = 2104 // Kafka消息发送失败
	KAFKA_MESSAGE_CONSUME_FAILED = 2105 // Kafka消息消费失败
	KAFKA_PARTITION_ERROR     = 2106 // Kafka分区错误
	KAFKA_OFFSET_ERROR        = 2107 // Kafka偏移量错误

	// ========== 缓存相关错误 (2200-2299) ==========
	REDIS_CONNECTION_FAILED   = 2200 // Redis连接失败
	REDIS_OPERATION_FAILED    = 2201 // Redis操作失败
	REDIS_KEY_NOT_FOUND       = 2202 // Redis键不存在
	REDIS_KEY_EXPIRED         = 2203 // Redis键已过期
	REDIS_MEMORY_FULL         = 2204 // Redis内存已满
	REDIS_CLUSTER_ERROR       = 2205 // Redis集群错误
	CACHE_MISS                = 2206 // 缓存未命中
	CACHE_INVALIDATION_FAILED = 2207 // 缓存失效失败

	// ========== 系统相关错误 (2300-2399) ==========
	SYSTEM_ERROR              = 2300 // 系统错误
	SERVICE_UNAVAILABLE       = 2301 // 服务不可用
	RATE_LIMIT_EXCEEDED       = 2302 // 请求频率超限
	RESOURCE_EXHAUSTED        = 2303 // 资源耗尽
	CONFIG_ERROR              = 2304 // 配置错误
	INTERNAL_ERROR            = 2305 // 内部错误
	TIMEOUT_ERROR             = 2306 // 超时错误
	CONCURRENCY_ERROR         = 2307 // 并发错误
	SHARD_NOT_FOUND           = 2308 // 分片未找到
	SHARD_OVERLOADED          = 2309 // 分片过载
	SHARD_STOPPED             = 2310 // 分片已停止
	WORKER_POOL_FULL          = 2311 // 工作池已满
	WORKER_NOT_AVAILABLE      = 2312 // 工作者不可用
	CIRCUIT_BREAKER_OPEN      = 2313 // 熔断器开启
	HEALTH_CHECK_FAILED       = 2314 // 健康检查失败
)
