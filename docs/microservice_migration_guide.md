# LionChat 微服务化迁移指南

## 概述

本指南详细说明如何将 LionChat 从当前的单体架构拆分为微服务架构。我们将采用渐进式迁移策略，确保系统在迁移过程中保持稳定运行。

## 迁移策略

### 1. 绞杀者模式 (Strangler Fig Pattern)

我们将采用绞杀者模式逐步替换单体应用的功能：

```
单体应用 → API Gateway → 微服务
    ↓           ↓           ↓
  保留      路由分发     新功能
```

### 2. 数据库分离策略

**阶段1: 数据库分离**
- 为每个微服务创建独立的数据库
- 使用数据同步工具保持数据一致性
- 逐步迁移数据访问逻辑

**阶段2: 服务拆分**
- 提取业务逻辑到独立服务
- 实现服务间通信
- 替换单体应用的对应功能

## 服务拆分详细方案

### 1. 用户服务 (User Service)

#### 1.1 服务职责
- 用户注册、登录、注销
- 用户信息管理
- 用户状态管理
- 头像上传和管理

#### 1.2 数据模型

```sql
-- 用户服务数据库
CREATE DATABASE user_service;

-- 用户表
CREATE TABLE users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    uuid VARCHAR(150) UNIQUE NOT NULL,
    username VARCHAR(150) UNIQUE NOT NULL,
    nickname VARCHAR(150),
    email VARCHAR(80),
    password VARCHAR(150) NOT NULL,
    avatar VARCHAR(150),
    status TINYINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- 用户配置表
CREATE TABLE user_settings (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    setting_key VARCHAR(100) NOT NULL,
    setting_value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    UNIQUE KEY uk_user_setting (user_id, setting_key)
);
```

#### 1.3 API 设计

```go
// internal/service/user/api.go
package user

import (
    "github.com/gin-gonic/gin"
)

type UserAPI struct {
    service *UserService
}

func NewUserAPI(service *UserService) *UserAPI {
    return &UserAPI{service: service}
}

func (api *UserAPI) RegisterRoutes(r *gin.RouterGroup) {
    users := r.Group("/users")
    {
        users.POST("/register", api.Register)
        users.POST("/login", api.Login)
        users.POST("/logout", api.Logout)
        users.GET("/:id", api.GetUser)
        users.PUT("/:id", api.UpdateUser)
        users.DELETE("/:id", api.DeleteUser)
        users.POST("/:id/avatar", api.UploadAvatar)
        users.GET("/:id/settings", api.GetUserSettings)
        users.PUT("/:id/settings", api.UpdateUserSettings)
    }
}

// 用户注册
func (api *UserAPI) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    user, err := api.service.Register(&req)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(201, gin.H{"data": user})
}

// 用户登录
func (api *UserAPI) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    token, err := api.service.Login(&req)
    if err != nil {
        c.JSON(401, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"token": token})
}
```

#### 1.4 服务实现

```go
// internal/service/user/service.go
package user

import (
    "context"
    "errors"
    "time"
    
    "gorm.io/gorm"
    "github.com/go-redis/redis/v8"
)

type UserService struct {
    db    *gorm.DB
    redis *redis.Client
}

func NewUserService(db *gorm.DB, redis *redis.Client) *UserService {
    return &UserService{
        db:    db,
        redis: redis,
    }
}

func (s *UserService) Register(req *RegisterRequest) (*User, error) {
    // 检查用户名是否已存在
    var existingUser User
    if err := s.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
        return nil, errors.New("用户名已存在")
    }
    
    // 创建新用户
    user := &User{
        UUID:     generateUUID(),
        Username: req.Username,
        Nickname: req.Nickname,
        Email:    req.Email,
        Password: hashPassword(req.Password),
        Status:   0,
    }
    
    if err := s.db.Create(user).Error; err != nil {
        return nil, err
    }
    
    // 发送用户注册事件
    s.publishUserEvent("user_registered", user.UUID)
    
    return user, nil
}

func (s *UserService) Login(req *LoginRequest) (string, error) {
    var user User
    if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
        return "", errors.New("用户不存在")
    }
    
    if !verifyPassword(req.Password, user.Password) {
        return "", errors.New("密码错误")
    }
    
    // 生成JWT令牌
    token, err := generateJWT(user.UUID)
    if err != nil {
        return "", err
    }
    
    // 更新用户在线状态
    s.redis.Set(context.Background(), "user:online:"+user.UUID, "1", time.Hour*24)
    
    // 发送用户登录事件
    s.publishUserEvent("user_login", user.UUID)
    
    return token, nil
}

func (s *UserService) publishUserEvent(eventType, userID string) {
    // 发送事件到Kafka
    event := UserEvent{
        EventType: eventType,
        UserID:    userID,
        Timestamp: time.Now().Unix(),
    }
    
    // 这里应该调用Kafka生产者
    // kafkaProducer.SendUserEvent(event)
}
```

### 2. 认证服务 (Auth Service)

#### 2.1 服务职责
- JWT 令牌生成和验证
- 权限管理
- 会话管理
- 单点登录 (SSO)

#### 2.2 服务实现

```go
// internal/service/auth/service.go
package auth

import (
    "context"
    "time"
    
    "github.com/golang-jwt/jwt/v4"
    "github.com/go-redis/redis/v8"
)

type AuthService struct {
    redis     *redis.Client
    jwtSecret string
}

func NewAuthService(redis *redis.Client, jwtSecret string) *AuthService {
    return &AuthService{
        redis:     redis,
        jwtSecret: jwtSecret,
    }
}

type Claims struct {
    UserID string `json:"user_id"`
    jwt.RegisteredClaims
}

func (s *AuthService) GenerateToken(userID string) (string, error) {
    claims := &Claims{
        UserID: userID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "lionchat",
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(s.jwtSecret))
    if err != nil {
        return "", err
    }
    
    // 将令牌存储到Redis中
    s.redis.Set(context.Background(), "token:"+userID, tokenString, 24*time.Hour)
    
    return tokenString, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(s.jwtSecret), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        // 检查令牌是否在Redis中存在
        exists, err := s.redis.Exists(context.Background(), "token:"+claims.UserID).Result()
        if err != nil || exists == 0 {
            return nil, errors.New("令牌已失效")
        }
        
        return claims, nil
    }
    
    return nil, errors.New("无效的令牌")
}

func (s *AuthService) RevokeToken(userID string) error {
    return s.redis.Del(context.Background(), "token:"+userID).Err()
}
```

### 3. 聊天服务 (Chat Service)

#### 3.1 服务职责
- WebSocket 连接管理
- 消息路由和分发
- 在线状态管理
- 消息持久化

#### 3.2 WebSocket 管理器

```go
// internal/service/chat/websocket.go
package chat

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "sync"
    
    "github.com/gorilla/websocket"
    "github.com/go-redis/redis/v8"
)

type WebSocketManager struct {
    connections map[string]*websocket.Conn
    mutex       sync.RWMutex
    redis       *redis.Client
    upgrader    websocket.Upgrader
    nodeID      string
}

func NewWebSocketManager(redis *redis.Client, nodeID string) *WebSocketManager {
    return &WebSocketManager{
        connections: make(map[string]*websocket.Conn),
        redis:       redis,
        nodeID:      nodeID,
        upgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool {
                return true // 在生产环境中应该进行适当的检查
            },
        },
    }
}

func (wsm *WebSocketManager) HandleConnection(w http.ResponseWriter, r *http.Request) {
    conn, err := wsm.upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket升级失败: %v", err)
        return
    }
    defer conn.Close()
    
    // 从查询参数或头部获取用户ID
    userID := r.URL.Query().Get("user_id")
    if userID == "" {
        log.Printf("缺少用户ID")
        return
    }
    
    // 注册连接
    wsm.registerConnection(userID, conn)
    defer wsm.unregisterConnection(userID)
    
    // 处理消息
    for {
        var message Message
        err := conn.ReadJSON(&message)
        if err != nil {
            log.Printf("读取消息失败: %v", err)
            break
        }
        
        wsm.handleMessage(userID, &message)
    }
}

func (wsm *WebSocketManager) registerConnection(userID string, conn *websocket.Conn) {
    wsm.mutex.Lock()
    defer wsm.mutex.Unlock()
    
    wsm.connections[userID] = conn
    
    // 在Redis中记录用户连接的节点
    wsm.redis.HSet(context.Background(), "user:connections", userID, wsm.nodeID)
    wsm.redis.Expire(context.Background(), "user:connections", 30*time.Minute)
    
    log.Printf("用户 %s 已连接到节点 %s", userID, wsm.nodeID)
}

func (wsm *WebSocketManager) unregisterConnection(userID string) {
    wsm.mutex.Lock()
    defer wsm.mutex.Unlock()
    
    delete(wsm.connections, userID)
    
    // 从Redis中移除用户连接记录
    wsm.redis.HDel(context.Background(), "user:connections", userID)
    
    log.Printf("用户 %s 已断开连接", userID)
}

func (wsm *WebSocketManager) handleMessage(senderID string, message *Message) {
    // 验证消息
    if err := wsm.validateMessage(message); err != nil {
        log.Printf("消息验证失败: %v", err)
        return
    }
    
    // 设置发送者ID
    message.SenderID = senderID
    message.Timestamp = time.Now().Unix()
    
    // 根据消息类型处理
    switch message.Type {
    case "private":
        wsm.handlePrivateMessage(message)
    case "group":
        wsm.handleGroupMessage(message)
    default:
        log.Printf("未知的消息类型: %s", message.Type)
    }
}

func (wsm *WebSocketManager) handlePrivateMessage(message *Message) {
    // 查找接收者连接的节点
    nodeID, err := wsm.redis.HGet(context.Background(), "user:connections", message.ReceiverID).Result()
    if err != nil {
        log.Printf("用户 %s 不在线", message.ReceiverID)
        // 可以选择将消息存储到离线消息队列
        return
    }
    
    if nodeID == wsm.nodeID {
        // 本地连接，直接发送
        wsm.sendToLocalConnection(message.ReceiverID, message)
    } else {
        // 远程连接，通过消息队列转发
        wsm.forwardToRemoteNode(nodeID, message)
    }
    
    // 持久化消息
    wsm.persistMessage(message)
}

func (wsm *WebSocketManager) sendToLocalConnection(userID string, message *Message) {
    wsm.mutex.RLock()
    conn, exists := wsm.connections[userID]
    wsm.mutex.RUnlock()
    
    if !exists {
        log.Printf("本地连接不存在: %s", userID)
        return
    }
    
    if err := conn.WriteJSON(message); err != nil {
        log.Printf("发送消息失败: %v", err)
        wsm.unregisterConnection(userID)
    }
}

func (wsm *WebSocketManager) forwardToRemoteNode(nodeID string, message *Message) {
    // 通过Kafka发送到远程节点
    forwardMessage := ForwardMessage{
        TargetNodeID: nodeID,
        Message:      message,
    }
    
    data, _ := json.Marshal(forwardMessage)
    // kafkaProducer.SendMessage("message-forward", nodeID, data)
}
```

### 4. 好友服务 (Friend Service)

#### 4.1 数据模型

```sql
-- 好友服务数据库
CREATE DATABASE friend_service;

-- 好友关系表
CREATE TABLE friendships (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(150) NOT NULL,
    friend_id VARCHAR(150) NOT NULL,
    status ENUM('pending', 'accepted', 'blocked') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_friendship (user_id, friend_id),
    INDEX idx_user_id (user_id),
    INDEX idx_friend_id (friend_id)
);

-- 好友申请表
CREATE TABLE friend_requests (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    from_user_id VARCHAR(150) NOT NULL,
    to_user_id VARCHAR(150) NOT NULL,
    message TEXT,
    status ENUM('pending', 'accepted', 'rejected') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_from_user (from_user_id),
    INDEX idx_to_user (to_user_id)
);
```

#### 4.2 服务实现

```go
// internal/service/friend/service.go
package friend

import (
    "context"
    "errors"
    "time"
    
    "gorm.io/gorm"
    "github.com/go-redis/redis/v8"
)

type FriendService struct {
    db    *gorm.DB
    redis *redis.Client
}

func NewFriendService(db *gorm.DB, redis *redis.Client) *FriendService {
    return &FriendService{
        db:    db,
        redis: redis,
    }
}

func (s *FriendService) SendFriendRequest(fromUserID, toUserID, message string) error {
    // 检查是否已经是好友
    var friendship Friendship
    if err := s.db.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)", 
        fromUserID, toUserID, toUserID, fromUserID).First(&friendship).Error; err == nil {
        return errors.New("已经是好友关系")
    }
    
    // 检查是否已经发送过申请
    var request FriendRequest
    if err := s.db.Where("from_user_id = ? AND to_user_id = ? AND status = 'pending'", 
        fromUserID, toUserID).First(&request).Error; err == nil {
        return errors.New("已经发送过好友申请")
    }
    
    // 创建好友申请
    request = FriendRequest{
        FromUserID: fromUserID,
        ToUserID:   toUserID,
        Message:    message,
        Status:     "pending",
    }
    
    if err := s.db.Create(&request).Error; err != nil {
        return err
    }
    
    // 发送通知事件
    s.publishFriendEvent("friend_request_sent", fromUserID, toUserID)
    
    return nil
}

func (s *FriendService) AcceptFriendRequest(requestID int64, userID string) error {
    var request FriendRequest
    if err := s.db.Where("id = ? AND to_user_id = ? AND status = 'pending'", 
        requestID, userID).First(&request).Error; err != nil {
        return errors.New("好友申请不存在")
    }
    
    // 开始事务
    tx := s.db.Begin()
    
    // 更新申请状态
    if err := tx.Model(&request).Update("status", "accepted").Error; err != nil {
        tx.Rollback()
        return err
    }
    
    // 创建双向好友关系
    friendships := []Friendship{
        {
            UserID:   request.FromUserID,
            FriendID: request.ToUserID,
            Status:   "accepted",
        },
        {
            UserID:   request.ToUserID,
            FriendID: request.FromUserID,
            Status:   "accepted",
        },
    }
    
    if err := tx.Create(&friendships).Error; err != nil {
        tx.Rollback()
        return err
    }
    
    tx.Commit()
    
    // 清除缓存
    s.clearFriendCache(request.FromUserID)
    s.clearFriendCache(request.ToUserID)
    
    // 发送通知事件
    s.publishFriendEvent("friend_request_accepted", request.FromUserID, request.ToUserID)
    
    return nil
}

func (s *FriendService) GetFriendList(userID string) ([]Friend, error) {
    // 先从缓存获取
    cacheKey := "friends:" + userID
    cached, err := s.redis.Get(context.Background(), cacheKey).Result()
    if err == nil {
        var friends []Friend
        json.Unmarshal([]byte(cached), &friends)
        return friends, nil
    }
    
    // 从数据库获取
    var friendships []Friendship
    if err := s.db.Where("user_id = ? AND status = 'accepted'", userID).Find(&friendships).Error; err != nil {
        return nil, err
    }
    
    var friends []Friend
    for _, friendship := range friendships {
        // 这里需要调用用户服务获取好友信息
        friend, err := s.getUserInfo(friendship.FriendID)
        if err != nil {
            continue
        }
        friends = append(friends, *friend)
    }
    
    // 缓存结果
    data, _ := json.Marshal(friends)
    s.redis.Set(context.Background(), cacheKey, data, time.Hour)
    
    return friends, nil
}

func (s *FriendService) clearFriendCache(userID string) {
    s.redis.Del(context.Background(), "friends:"+userID)
}

func (s *FriendService) getUserInfo(userID string) (*Friend, error) {
    // 这里应该调用用户服务的API
    // 可以使用HTTP客户端或gRPC客户端
    // 为了简化，这里返回模拟数据
    return &Friend{
        UserID:   userID,
        Username: "user_" + userID,
        Nickname: "User " + userID,
        Avatar:   "/avatars/default.png",
    }, nil
}

func (s *FriendService) publishFriendEvent(eventType, fromUserID, toUserID string) {
    event := FriendEvent{
        EventType:  eventType,
        FromUserID: fromUserID,
        ToUserID:   toUserID,
        Timestamp:  time.Now().Unix(),
    }
    
    // 发送到Kafka
    // kafkaProducer.SendFriendEvent(event)
}
```

## 服务间通信

### 1. 同步通信 (HTTP/gRPC)

```go
// pkg/client/user_client.go
package client

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type UserClient struct {
    baseURL    string
    httpClient *http.Client
}

func NewUserClient(baseURL string) *UserClient {
    return &UserClient{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (c *UserClient) GetUser(userID string) (*User, error) {
    url := fmt.Sprintf("%s/api/users/%s", c.baseURL, userID)
    
    resp, err := c.httpClient.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("获取用户信息失败: %d", resp.StatusCode)
    }
    
    var response struct {
        Data User `json:"data"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, err
    }
    
    return &response.Data, nil
}

func (c *UserClient) ValidateUser(userID string) (bool, error) {
    url := fmt.Sprintf("%s/api/users/%s/validate", c.baseURL, userID)
    
    resp, err := c.httpClient.Get(url)
    if err != nil {
        return false, err
    }
    defer resp.Body.Close()
    
    return resp.StatusCode == http.StatusOK, nil
}
```

### 2. 异步通信 (Kafka)

```go
// pkg/events/event_handler.go
package events

import (
    "encoding/json"
    "log"
    
    "github.com/IBM/sarama"
)

type EventHandler struct {
    userService   UserService
    friendService FriendService
    chatService   ChatService
}

func NewEventHandler(userService UserService, friendService FriendService, chatService ChatService) *EventHandler {
    return &EventHandler{
        userService:   userService,
        friendService: friendService,
        chatService:   chatService,
    }
}

func (h *EventHandler) HandleMessage(message *sarama.ConsumerMessage) {
    var event Event
    if err := json.Unmarshal(message.Value, &event); err != nil {
        log.Printf("解析事件失败: %v", err)
        return
    }
    
    switch event.Type {
    case "user_registered":
        h.handleUserRegistered(&event)
    case "friend_request_sent":
        h.handleFriendRequestSent(&event)
    case "message_sent":
        h.handleMessageSent(&event)
    default:
        log.Printf("未知事件类型: %s", event.Type)
    }
}

func (h *EventHandler) handleUserRegistered(event *Event) {
    // 为新用户创建默认设置
    // 发送欢迎消息等
    log.Printf("处理用户注册事件: %s", event.UserID)
}

func (h *EventHandler) handleFriendRequestSent(event *Event) {
    // 发送推送通知
    // 更新未读计数等
    log.Printf("处理好友申请事件: %s -> %s", event.FromUserID, event.ToUserID)
}

func (h *EventHandler) handleMessageSent(event *Event) {
    // 更新会话列表
    // 发送推送通知
    // 更新未读计数
    log.Printf("处理消息发送事件: %s", event.MessageID)
}
```

## 数据一致性策略

### 1. 最终一致性

对于大部分业务场景，我们采用最终一致性模型：

```go
// pkg/consistency/saga.go
package consistency

import (
    "context"
    "encoding/json"
    "time"
)

// Saga 模式实现分布式事务
type Saga struct {
    ID          string
    Steps       []SagaStep
    Status      string
    CreatedAt   time.Time
    CompletedAt *time.Time
}

type SagaStep struct {
    ID            string
    ServiceName   string
    Action        string
    CompensateAction string
    Payload       json.RawMessage
    Status        string
    ExecutedAt    *time.Time
}

type SagaManager struct {
    storage SagaStorage
    eventBus EventBus
}

func (sm *SagaManager) ExecuteSaga(saga *Saga) error {
    // 保存Saga状态
    if err := sm.storage.SaveSaga(saga); err != nil {
        return err
    }
    
    // 执行第一步
    return sm.executeNextStep(saga)
}

func (sm *SagaManager) executeNextStep(saga *Saga) error {
    for i, step := range saga.Steps {
        if step.Status == "pending" {
            // 发送执行命令
            command := SagaCommand{
                SagaID:      saga.ID,
                StepID:      step.ID,
                ServiceName: step.ServiceName,
                Action:      step.Action,
                Payload:     step.Payload,
            }
            
            return sm.eventBus.PublishCommand(&command)
        }
    }
    
    // 所有步骤都已完成
    saga.Status = "completed"
    now := time.Now()
    saga.CompletedAt = &now
    
    return sm.storage.SaveSaga(saga)
}

func (sm *SagaManager) HandleStepCompleted(sagaID, stepID string, success bool) error {
    saga, err := sm.storage.GetSaga(sagaID)
    if err != nil {
        return err
    }
    
    // 更新步骤状态
    for i, step := range saga.Steps {
        if step.ID == stepID {
            if success {
                saga.Steps[i].Status = "completed"
                now := time.Now()
                saga.Steps[i].ExecutedAt = &now
            } else {
                saga.Steps[i].Status = "failed"
                // 开始补偿操作
                return sm.startCompensation(saga, i)
            }
            break
        }
    }
    
    // 保存状态并执行下一步
    if err := sm.storage.SaveSaga(saga); err != nil {
        return err
    }
    
    return sm.executeNextStep(saga)
}

func (sm *SagaManager) startCompensation(saga *Saga, failedStepIndex int) error {
    saga.Status = "compensating"
    
    // 从失败步骤开始，逆序执行补偿操作
    for i := failedStepIndex - 1; i >= 0; i-- {
        step := saga.Steps[i]
        if step.Status == "completed" {
            compensateCommand := SagaCommand{
                SagaID:      saga.ID,
                StepID:      step.ID,
                ServiceName: step.ServiceName,
                Action:      step.CompensateAction,
                Payload:     step.Payload,
            }
            
            if err := sm.eventBus.PublishCommand(&compensateCommand); err != nil {
                return err
            }
        }
    }
    
    saga.Status = "compensated"
    return sm.storage.SaveSaga(saga)
}
```

### 2. 事件溯源

```go
// pkg/eventsourcing/event_store.go
package eventsourcing

import (
    "encoding/json"
    "time"
)

type Event struct {
    ID            string          `json:"id"`
    AggregateID   string          `json:"aggregate_id"`
    AggregateType string          `json:"aggregate_type"`
    EventType     string          `json:"event_type"`
    EventData     json.RawMessage `json:"event_data"`
    Version       int64           `json:"version"`
    Timestamp     time.Time       `json:"timestamp"`
}

type EventStore interface {
    SaveEvents(aggregateID string, events []Event, expectedVersion int64) error
    GetEvents(aggregateID string, fromVersion int64) ([]Event, error)
    GetAllEvents(fromTimestamp time.Time) ([]Event, error)
}

type Aggregate interface {
    GetID() string
    GetVersion() int64
    GetUncommittedEvents() []Event
    MarkEventsAsCommitted()
    LoadFromHistory(events []Event)
}

// 用户聚合根示例
type UserAggregate struct {
    ID               string
    Version          int64
    Username         string
    Email            string
    Status           string
    uncommittedEvents []Event
}

func (u *UserAggregate) Register(username, email, password string) {
    event := Event{
        ID:            generateEventID(),
        AggregateID:   u.ID,
        AggregateType: "User",
        EventType:     "UserRegistered",
        EventData:     marshalEventData(UserRegisteredEvent{
            Username: username,
            Email:    email,
            Password: password,
        }),
        Version:   u.Version + 1,
        Timestamp: time.Now(),
    }
    
    u.apply(event)
    u.uncommittedEvents = append(u.uncommittedEvents, event)
}

func (u *UserAggregate) apply(event Event) {
    switch event.EventType {
    case "UserRegistered":
        var data UserRegisteredEvent
        json.Unmarshal(event.EventData, &data)
        u.Username = data.Username
        u.Email = data.Email
        u.Status = "active"
    case "UserDeactivated":
        u.Status = "inactive"
    }
    
    u.Version = event.Version
}

func (u *UserAggregate) LoadFromHistory(events []Event) {
    for _, event := range events {
        u.apply(event)
    }
    u.uncommittedEvents = nil
}
```

## 迁移步骤

### 阶段1: 准备工作 (1周)

1. **创建新的代码仓库结构**
```
lionchat-microservices/
├── services/
│   ├── user-service/
│   ├── auth-service/
│   ├── chat-service/
│   ├── friend-service/
│   ├── group-service/
│   └── file-service/
├── pkg/
│   ├── client/
│   ├── events/
│   ├── middleware/
│   └── utils/
├── deployments/
│   ├── docker/
│   ├── kubernetes/
│   └── terraform/
└── scripts/
    ├── migration/
    └── deployment/
```

2. **设置CI/CD流水线**
3. **准备测试环境**

### 阶段2: 用户服务拆分 (1-2周)

1. **创建用户服务数据库**
2. **实现用户服务API**
3. **数据迁移脚本**
4. **集成测试**

### 阶段3: 认证服务拆分 (1周)

1. **实现JWT服务**
2. **会话管理**
3. **权限验证中间件**

### 阶段4: 聊天服务拆分 (2-3周)

1. **WebSocket连接管理**
2. **消息路由逻辑**
3. **跨节点消息转发**
4. **性能测试**

### 阶段5: 其他服务拆分 (2-3周)

1. **好友服务**
2. **群组服务**
3. **文件服务**
4. **通知服务**

### 阶段6: 集成和优化 (1-2周)

1. **服务间通信优化**
2. **性能调优**
3. **监控和告警**
4. **文档完善**

## 风险控制

### 1. 回滚策略

- 保持原有单体应用可用
- 使用特性开关控制流量切换
- 准备快速回滚脚本

### 2. 灰度发布

```go
// 流量分流中间件
func TrafficSplitMiddleware(percentage int) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        hash := hashUserID(userID)
        
        if hash%100 < percentage {
            // 路由到新服务
            c.Set("use_microservice", true)
        } else {
            // 路由到原有服务
            c.Set("use_microservice", false)
        }
        
        c.Next()
    }
}
```

### 3. 数据一致性检查

```go
// 数据一致性检查工具
func CheckDataConsistency() error {
    // 比较单体应用和微服务的数据
    // 发现不一致时发出告警
    return nil
}
```

## 总结

通过这个详细的迁移指南，我们可以安全、有序地将LionChat从单体架构迁移到微服务架构。关键要点：

1. **渐进式迁移**: 避免大爆炸式重写
2. **数据一致性**: 使用事件驱动和最终一致性
3. **服务治理**: 完善的监控、日志和追踪
4. **风险控制**: 灰度发布和快速回滚能力
5. **团队协作**: 清晰的服务边界和接口定义

这种方法可以最大程度地降低迁移风险，确保系统在迁移过程中保持稳定运行。