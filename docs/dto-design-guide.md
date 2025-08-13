# LionChat DTO 设计规范指南

## 概述

本文档详细说明了 LionChat 项目中数据传输对象（DTO）的设计规范和最佳实践。我们的 DTO 设计参考了业界著名开源项目（如 GitHub API、GitLab API、Twitter API 等）的最佳实践，旨在提供一致、可维护、可扩展的 API 数据结构。

## 设计原则

### 1. 统一性原则
- 所有 API 响应使用统一的响应结构
- 请求和响应的命名遵循一致的规范
- 错误处理采用标准化的格式

### 2. 可扩展性原则
- 使用泛型支持类型安全
- 预留扩展字段，便于未来功能迭代
- 支持版本化的 API 设计

### 3. 安全性原则
- 输入验证在 DTO 层面进行
- 敏感信息不在响应中暴露
- 支持权限控制字段

### 4. 性能原则
- 区分详细信息和摘要信息
- 支持分页和游标分页
- 避免 N+1 查询问题

## 核心结构设计

### 1. 基础响应结构

#### APIResponse[T]
```go
type APIResponse[T any] struct {
    Code      int    `json:"code"`                // 业务状态码
    Message   string `json:"message"`             // 响应消息
    Data      T      `json:"data,omitempty"`      // 响应数据
    Timestamp int64  `json:"timestamp"`           // 时间戳
    RequestID string `json:"request_id,omitempty"` // 请求ID，用于追踪
}
```

**设计说明：**
- 使用泛型 `T` 确保类型安全
- `Code` 字段用于业务状态码，与 HTTP 状态码分离
- `RequestID` 用于请求追踪和调试
- 参考了 GitHub API 和 GitLab API 的响应结构

#### PagedResponse[T]
```go
type PagedResponse[T any] struct {
    Code      int      `json:"code"`
    Message   string   `json:"message"`
    Data      []T      `json:"data"`
    PageInfo  PageInfo `json:"page_info"`
    Timestamp int64    `json:"timestamp"`
    RequestID string   `json:"request_id,omitempty"`
}
```

**设计说明：**
- 专门用于分页数据的响应
- `PageInfo` 包含分页元信息
- 参考了 GitHub API 的分页设计

### 2. 错误处理结构

#### ErrorResponse
```go
type ErrorResponse struct {
    Code      int           `json:"code"`
    Message   string        `json:"message"`
    Errors    []ErrorDetail `json:"errors,omitempty"`
    Timestamp int64         `json:"timestamp"`
    RequestID string        `json:"request_id,omitempty"`
}
```

**设计说明：**
- 遵循 RFC 7807 Problem Details 标准
- 支持多个错误详情
- 便于客户端进行错误处理

### 3. 基础请求结构

#### BaseRequest
```go
type BaseRequest struct {
    RequestID string `json:"request_id,omitempty" validate:"-"`
}
```

**设计说明：**
- 所有请求 DTO 都应嵌入此结构
- 提供请求追踪能力

### 4. 混入结构（Mixin）

#### TimestampMixin
```go
type TimestampMixin struct {
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
```

#### UserMixin
```go
type UserMixin struct {
    UserID   int64  `json:"user_id"`
    Username string `json:"username"`
    Nickname string `json:"nickname"`
    Avatar   string `json:"avatar"`
}
```

**设计说明：**
- 使用组合模式避免代码重复
- 提供一致的时间戳和用户信息结构

## 验证规范

### 1. 验证标签使用

我们使用 `go-playground/validator` 库进行输入验证：

```go
type LoginRequest struct {
    BaseRequest
    Username string `json:"username" validate:"required,min=3,max=50" example:"john_doe"`
    Password string `json:"password" validate:"required,min=6,max=128" example:"password123"`
}
```

### 2. 常用验证规则

| 规则 | 说明 | 示例 |
|------|------|------|
| `required` | 必填字段 | `validate:"required"` |
| `min=n,max=n` | 字符串长度限制 | `validate:"min=3,max=50"` |
| `email` | 邮箱格式 | `validate:"email"` |
| `url` | URL 格式 | `validate:"url"` |
| `oneof` | 枚举值 | `validate:"oneof=public private"` |
| `dive` | 数组元素验证 | `validate:"dive,required"` |

## 命名规范

### 1. 请求 DTO 命名

- 创建：`Create{Entity}Request`
- 更新：`Update{Entity}Request`
- 删除：`Delete{Entity}Request`
- 查询：`{Entity}ListRequest`、`{Entity}SearchRequest`
- 详情：`{Entity}DetailRequest`

### 2. 响应 DTO 命名

- 详细信息：`{Entity}Profile`
- 摘要信息：`{Entity}Summary`
- 统计信息：`{Entity}Stats`
- 设置信息：`{Entity}Settings`

### 3. 字段命名

- 使用 `snake_case` 的 JSON 标签
- 布尔字段使用 `is_` 或 `can_` 前缀
- 时间字段使用 `_at` 后缀
- 计数字段使用 `_count` 后缀

## 分页设计

### 1. 传统分页

```go
type PageInfo struct {
    Page     int `json:"page"`      // 当前页码
    PageSize int `json:"page_size"` // 每页大小
    Total    int `json:"total"`     // 总记录数
    Pages    int `json:"pages"`     // 总页数
}
```

### 2. 游标分页

```go
type MomentListRequest struct {
    SinceID *int64 `json:"since_id"` // 获取指定ID之后的数据
    MaxID   *int64 `json:"max_id"`   // 获取指定ID之前的数据
    Limit   int    `json:"limit"`    // 限制数量
}
```

**使用场景：**
- 传统分页：适用于需要跳页的场景
- 游标分页：适用于实时数据流，如时间线

## 权限控制字段

在响应 DTO 中包含权限控制字段，便于前端进行 UI 控制：

```go
type MomentProfile struct {
    // ... 其他字段
    CanEdit   bool `json:"can_edit"`   // 当前用户是否可编辑
    CanDelete bool `json:"can_delete"` // 当前用户是否可删除
}
```

## 状态码设计

### 1. 业务状态码

| 状态码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1001-1999 | 用户相关错误 |
| 2001-2999 | 好友相关错误 |
| 3001-3999 | 群组相关错误 |
| 4001-4999 | 动态相关错误 |
| 5001-5999 | 评论相关错误 |
| 9001-9999 | 系统错误 |

### 2. HTTP 状态码

- 200：成功
- 400：请求参数错误
- 401：未授权
- 403：权限不足
- 404：资源不存在
- 500：服务器内部错误

## 最佳实践

### 1. 数据脱敏

```go
type UserProfile struct {
    Email string `json:"email,omitempty"` // 根据权限决定是否返回
    Phone string `json:"phone,omitempty"` // 敏感信息可选返回
}
```

### 2. 版本兼容

```go
type UserProfileV2 struct {
    UserProfile
    Bio      string `json:"bio,omitempty"`      // 新增字段
    Location string `json:"location,omitempty"` // 新增字段
}
```

### 3. 性能优化

- 使用 `Summary` 类型减少数据传输
- 避免在列表接口返回过多详细信息
- 合理使用 `omitempty` 标签

### 4. 文档化

```go
type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50" example:"john_doe"`
    // 使用 example 标签提供示例值
}
```

## 参考项目

我们的 DTO 设计参考了以下著名开源项目：

1. **GitHub API v4**：响应结构设计、分页机制
2. **GitLab API**：错误处理、权限控制字段
3. **Twitter API v2**：游标分页、数据结构设计
4. **Stripe API**：请求验证、错误详情
5. **Slack API**：实时数据结构、用户信息设计

## 总结

通过采用这套 DTO 设计规范，我们实现了：

1. **一致性**：统一的请求响应结构
2. **可维护性**：清晰的命名规范和组织结构
3. **可扩展性**：支持未来功能迭代
4. **安全性**：完善的输入验证和权限控制
5. **性能**：合理的数据结构设计

这套规范将帮助团队构建更加健壮、易维护的 API 接口。