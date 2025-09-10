# LionChat 分布式部署方案

## 项目架构分析

### 当前架构组件
LionChat 是一个基于 Go 语言的即时通讯系统，当前采用单体架构，主要组件包括：

- **Web 服务层**: Gin 框架提供 HTTP API 和 WebSocket 连接
- **业务逻辑层**: 用户管理、消息处理、群组管理、好友管理等服务
- **数据存储层**: MySQL 数据库、Redis 缓存
- **消息队列**: Kafka 用于异步消息处理
- **AI 集成**: DeepSeek AI 聊天机器人
- **Worker 池**: 处理 WebSocket 连接和消息分发

### 技术栈
- **后端**: Go + Gin + GORM
- **数据库**: MySQL
- **缓存**: Redis
- **消息队列**: Kafka
- **容器化**: Docker + Docker Compose
- **WebSocket**: 原生 WebSocket 实现

## 分布式部署架构设计

### 1. 微服务拆分策略

#### 1.1 服务拆分原则
- **业务边界清晰**: 按业务功能拆分
- **数据独立**: 每个服务拥有独立的数据存储
- **松耦合**: 服务间通过 API 和消息队列通信
- **高内聚**: 相关功能聚合在同一服务内

#### 1.2 拆分后的微服务架构

```
┌─────────────────────────────────────────────────────────────┐
│                        API Gateway                          │
│                    (Nginx/Kong/Envoy)                      │
└─────────────────────────────────────────────────────────────┘
                                │
                ┌───────────────┼───────────────┐
                │               │               │
        ┌───────▼──────┐ ┌──────▼──────┐ ┌─────▼──────┐
        │ User Service │ │Auth Service │ │Chat Service│
        │              │ │             │ │            │
        │ - 用户管理    │ │ - JWT认证   │ │ - 消息处理  │
        │ - 个人资料    │ │ - 登录登出   │ │ - WebSocket │
        │ - 头像管理    │ │ - 权限验证   │ │ - 消息分发  │
        └──────────────┘ └─────────────┘ └────────────┘
                │               │               │
        ┌───────▼──────┐ ┌──────▼──────┐ ┌─────▼──────┐
        │Friend Service│ │Group Service│ │File Service│
        │              │ │             │ │            │
        │ - 好友管理    │ │ - 群组管理   │ │ - 文件上传  │
        │ - 好友申请    │ │ - 群成员管理 │ │ - 图片处理  │
        │ - 黑名单      │ │ - 群权限     │ │ - 文件存储  │
        └──────────────┘ └─────────────┘ └────────────┘
                │               │               │
        ┌───────▼──────┐ ┌──────▼──────┐ ┌─────▼──────┐
        │Moment Service│ │ Bot Service │ │Notify Svc  │
        │              │ │             │ │            │
        │ - 朋友圈      │ │ - AI聊天    │ │ - 推送通知  │
        │ - 动态发布    │ │ - DeepSeek  │ │ - 邮件通知  │
        │ - 评论点赞    │ │ - 智能回复   │ │ - 短信通知  │
        └──────────────┘ └─────────────┘ └────────────┘
```

### 2. 数据存储分布策略

#### 2.1 数据库分片策略

**用户数据分片 (User Sharding)**
```sql
-- 按用户ID哈希分片
Shard Key: user_id
Sharding Function: hash(user_id) % shard_count

-- 分片示例
user_shard_0: user_id % 4 = 0
user_shard_1: user_id % 4 = 1
user_shard_2: user_id % 4 = 2
user_shard_3: user_id % 4 = 3
```

**消息数据分片 (Message Sharding)**
```sql
-- 按会话ID分片，保证同一会话的消息在同一分片
Shard Key: conversation_id
Sharding Function: hash(conversation_id) % shard_count

-- 时间分片 (可选)
-- 按月份分表，便于历史数据归档
Table Pattern: messages_YYYYMM
```

#### 2.2 Redis 集群策略

```yaml
# Redis 集群配置
Redis Cluster:
  - Master-Slave 模式
  - 3 Master + 3 Slave 节点
  - 数据分片策略:
    - 用户在线状态: user:online:{user_id}
    - 用户会话: user:session:{user_id}
    - 消息缓存: msg:cache:{conversation_id}
    - 群组缓存: group:cache:{group_id}
```

### 3. 负载均衡和服务发现

#### 3.1 API Gateway 配置

```nginx
# Nginx 配置示例
upstream user_service {
    server user-service-1:8080 weight=1;
    server user-service-2:8080 weight=1;
    server user-service-3:8080 weight=1;
}

upstream chat_service {
    server chat-service-1:8081 weight=1;
    server chat-service-2:8081 weight=1;
    ip_hash; # WebSocket 连接保持
}

location /api/user/ {
    proxy_pass http://user_service;
}

location /api/chat/ {
    proxy_pass http://chat_service;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
}
```

#### 3.2 服务发现方案

**选项1: Consul + Consul Template**
```go
// 服务注册
func RegisterService(serviceName, serviceID, address string, port int) {
    client, _ := consul.NewClient(consul.DefaultConfig())
    
    registration := &consul.AgentServiceRegistration{
        ID:      serviceID,
        Name:    serviceName,
        Port:    port,
        Address: address,
        Check: &consul.AgentServiceCheck{
            HTTP:     fmt.Sprintf("http://%s:%d/health", address, port),
            Interval: "10s",
        },
    }
    
    client.Agent().ServiceRegister(registration)
}
```

**选项2: Kubernetes Service Discovery**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: user-service
spec:
  selector:
    app: user-service
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
```

### 4. 消息队列架构

#### 4.1 Kafka 集群配置

```yaml
# Kafka 集群部署
Kafka Cluster:
  Brokers: 3 节点
  Zookeeper: 3 节点
  Replication Factor: 3
  
Topics:
  - chat-messages: 12 partitions
  - user-events: 6 partitions
  - group-messages: 12 partitions
  - file-events: 3 partitions
  - notification-events: 6 partitions
```

#### 4.2 消息路由策略

```go
// 消息分区策略
type MessageRouter struct {
    producer sarama.SyncProducer
}

func (mr *MessageRouter) RouteMessage(msg *Message) {
    var partition int32
    
    switch msg.Type {
    case "private":
        // 私聊消息按会话ID分区
        partition = hash(msg.ConversationID) % 12
    case "group":
        // 群聊消息按群组ID分区
        partition = hash(msg.GroupID) % 12
    }
    
    mr.producer.SendMessage(&sarama.ProducerMessage{
        Topic:     "chat-messages",
        Partition: partition,
        Key:       sarama.StringEncoder(msg.ConversationID),
        Value:     sarama.ByteEncoder(msg.Data),
    })
}
```

### 5. WebSocket 连接管理

#### 5.1 连接分布策略

```go
// WebSocket 连接管理器
type ConnectionManager struct {
    connections map[string]*websocket.Conn
    nodeID      string
    redis       *redis.Client
}

// 用户连接时注册到Redis
func (cm *ConnectionManager) RegisterConnection(userID string, conn *websocket.Conn) {
    cm.connections[userID] = conn
    
    // 在Redis中记录用户连接的节点
    cm.redis.HSet(context.Background(), "user:connections", userID, cm.nodeID)
    cm.redis.Expire(context.Background(), "user:connections", 30*time.Minute)
}

// 消息路由到正确的节点
func (cm *ConnectionManager) RouteMessage(userID string, message []byte) {
    nodeID, err := cm.redis.HGet(context.Background(), "user:connections", userID).Result()
    if err != nil {
        return
    }
    
    if nodeID == cm.nodeID {
        // 本地连接，直接发送
        if conn, exists := cm.connections[userID]; exists {
            conn.WriteMessage(websocket.TextMessage, message)
        }
    } else {
        // 远程连接，通过消息队列转发
        cm.forwardToNode(nodeID, userID, message)
    }
}
```

### 6. 容器化部署方案

#### 6.1 Docker Compose 多环境配置

**开发环境 (docker-compose.dev.yml)**
```yaml
version: '3.8'
services:
  # 基础设施
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: lionchat
    volumes:
      - mysql_data:/var/lib/mysql
    ports:
      - "3306:3306"

  redis:
    image: redis:6.2-alpine
    ports:
      - "6379:6379"

  kafka:
    image: confluentinc/cp-kafka:7.4.0
    environment:
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
    ports:
      - "9092:9092"

  # 微服务
  user-service:
    build:
      context: ./services/user
    environment:
      - DB_HOST=mysql
      - REDIS_HOST=redis
      - KAFKA_BROKERS=kafka:9092
    depends_on:
      - mysql
      - redis
      - kafka

  chat-service:
    build:
      context: ./services/chat
    environment:
      - REDIS_HOST=redis
      - KAFKA_BROKERS=kafka:9092
    depends_on:
      - redis
      - kafka
    ports:
      - "8081:8081"

  api-gateway:
    image: nginx:alpine
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
    ports:
      - "80:80"
    depends_on:
      - user-service
      - chat-service
```

**生产环境 (docker-compose.prod.yml)**
```yaml
version: '3.8'
services:
  # 使用外部数据库和缓存
  user-service:
    image: lionchat/user-service:latest
    deploy:
      replicas: 3
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
    environment:
      - DB_HOST=${DB_HOST}
      - REDIS_HOST=${REDIS_HOST}
      - KAFKA_BROKERS=${KAFKA_BROKERS}
    networks:
      - lionchat-network

  chat-service:
    image: lionchat/chat-service:latest
    deploy:
      replicas: 3
      resources:
        limits:
          memory: 1G
          cpus: '1.0'
    environment:
      - REDIS_HOST=${REDIS_HOST}
      - KAFKA_BROKERS=${KAFKA_BROKERS}
    networks:
      - lionchat-network

networks:
  lionchat-network:
    external: true
```

#### 6.2 Kubernetes 部署配置

**用户服务部署**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
    spec:
      containers:
      - name: user-service
        image: lionchat/user-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: db.host
        - name: REDIS_HOST
          valueFrom:
            configMapKeyRef:
              name: app-config
              key: redis.host
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: user-service
spec:
  selector:
    app: user-service
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
```

### 7. 监控和日志方案

#### 7.1 监控架构

```yaml
# Prometheus + Grafana 监控栈
monitoring:
  prometheus:
    - 服务指标收集
    - 自定义业务指标
    - 告警规则配置
  
  grafana:
    - 可视化仪表板
    - 实时监控图表
    - 告警通知
  
  jaeger:
    - 分布式链路追踪
    - 性能分析
    - 错误定位
```

**关键监控指标**
```go
// 业务指标
var (
    // 消息相关
    messagesSent = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "messages_sent_total",
            Help: "Total number of messages sent",
        },
        []string{"type", "status"},
    )
    
    // 连接相关
    activeConnections = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "websocket_connections_active",
            Help: "Number of active WebSocket connections",
        },
    )
    
    // 响应时间
    responseTime = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration",
        },
        []string{"method", "endpoint", "status"},
    )
)
```

#### 7.2 日志收集方案

**ELK Stack 配置**
```yaml
# Filebeat 配置
filebeat.inputs:
- type: log
  enabled: true
  paths:
    - /app/logs/*.log
  fields:
    service: user-service
    environment: production
  fields_under_root: true

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "lionchat-logs-%{+yyyy.MM.dd}"

# Logstash 过滤器
filter {
  if [service] == "user-service" {
    grok {
      match => { "message" => "%{TIMESTAMP_ISO8601:timestamp} %{LOGLEVEL:level} %{GREEDYDATA:message}" }
    }
    
    date {
      match => [ "timestamp", "ISO8601" ]
    }
  }
}
```

### 8. 部署流程和CI/CD

#### 8.1 GitLab CI/CD 配置

```yaml
# .gitlab-ci.yml
stages:
  - build
  - test
  - deploy-staging
  - deploy-production

variables:
  DOCKER_REGISTRY: registry.gitlab.com/lionchat
  DOCKER_DRIVER: overlay2

build:
  stage: build
  script:
    - docker build -t $DOCKER_REGISTRY/user-service:$CI_COMMIT_SHA ./services/user
    - docker push $DOCKER_REGISTRY/user-service:$CI_COMMIT_SHA
  only:
    - main
    - develop

test:
  stage: test
  script:
    - go test ./...
    - go vet ./...
  coverage: '/coverage: \d+\.\d+% of statements/'

deploy-staging:
  stage: deploy-staging
  script:
    - kubectl set image deployment/user-service user-service=$DOCKER_REGISTRY/user-service:$CI_COMMIT_SHA
    - kubectl rollout status deployment/user-service
  environment:
    name: staging
    url: https://staging.lionchat.com
  only:
    - develop

deploy-production:
  stage: deploy-production
  script:
    - kubectl set image deployment/user-service user-service=$DOCKER_REGISTRY/user-service:$CI_COMMIT_SHA
    - kubectl rollout status deployment/user-service
  environment:
    name: production
    url: https://lionchat.com
  when: manual
  only:
    - main
```

### 9. 性能优化策略

#### 9.1 缓存策略

```go
// 多级缓存架构
type CacheManager struct {
    localCache  *bigcache.BigCache  // 本地缓存
    redisCache  *redis.Client      // Redis缓存
    mysqlDB     *gorm.DB           // 数据库
}

func (cm *CacheManager) GetUser(userID string) (*User, error) {
    // L1: 本地缓存
    if data, err := cm.localCache.Get("user:" + userID); err == nil {
        var user User
        json.Unmarshal(data, &user)
        return &user, nil
    }
    
    // L2: Redis缓存
    if data, err := cm.redisCache.Get(context.Background(), "user:"+userID).Result(); err == nil {
        var user User
        json.Unmarshal([]byte(data), &user)
        
        // 回写到本地缓存
        cm.localCache.Set("user:"+userID, []byte(data))
        return &user, nil
    }
    
    // L3: 数据库
    var user User
    if err := cm.mysqlDB.Where("id = ?", userID).First(&user).Error; err != nil {
        return nil, err
    }
    
    // 回写到缓存
    data, _ := json.Marshal(user)
    cm.redisCache.Set(context.Background(), "user:"+userID, data, time.Hour)
    cm.localCache.Set("user:"+userID, data)
    
    return &user, nil
}
```

#### 9.2 数据库优化

```sql
-- 索引优化
CREATE INDEX idx_messages_conversation_time ON messages(conversation_id, created_at);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_friends_user_friend ON friends(user_id, friend_id);

-- 分区表
CREATE TABLE messages (
    id BIGINT AUTO_INCREMENT,
    conversation_id VARCHAR(50),
    content TEXT,
    created_at TIMESTAMP,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (UNIX_TIMESTAMP(created_at)) (
    PARTITION p202401 VALUES LESS THAN (UNIX_TIMESTAMP('2024-02-01')),
    PARTITION p202402 VALUES LESS THAN (UNIX_TIMESTAMP('2024-03-01')),
    PARTITION p202403 VALUES LESS THAN (UNIX_TIMESTAMP('2024-04-01'))
);
```

### 10. 安全策略

#### 10.1 网络安全

```yaml
# 网络隔离
networks:
  frontend:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
  backend:
    driver: bridge
    internal: true
    ipam:
      config:
        - subnet: 172.21.0.0/16

# 服务间通信加密
services:
  api-gateway:
    networks:
      - frontend
      - backend
  
  user-service:
    networks:
      - backend
```

#### 10.2 认证授权

```go
// JWT 中间件
func JWTMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "Missing token"})
            c.Abort()
            return
        }
        
        claims, err := validateJWT(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}

// 权限控制
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("user_id")
        if !hasPermission(userID, permission) {
            c.JSON(403, gin.H{"error": "Insufficient permissions"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## 部署实施计划

### 阶段1: 基础设施准备 (1-2周)
1. 搭建 Kubernetes 集群
2. 配置 CI/CD 流水线
3. 部署监控和日志系统
4. 配置网络和安全策略

### 阶段2: 服务拆分 (3-4周)
1. 拆分用户服务
2. 拆分认证服务
3. 拆分聊天服务
4. 配置服务间通信

### 阶段3: 数据层改造 (2-3周)
1. 实施数据库分片
2. 配置 Redis 集群
3. 优化 Kafka 集群
4. 数据迁移和验证

### 阶段4: 性能优化 (1-2周)
1. 实施缓存策略
2. 优化数据库查询
3. 调整负载均衡
4. 性能测试和调优

### 阶段5: 生产部署 (1周)
1. 灰度发布
2. 全量切换
3. 监控和告警
4. 问题修复和优化

## 总结

这个分布式部署方案将 LionChat 从单体架构转换为微服务架构，具有以下优势：

1. **可扩展性**: 各服务可独立扩展
2. **可维护性**: 代码模块化，便于维护
3. **可靠性**: 服务隔离，故障不会影响整个系统
4. **性能**: 通过缓存、分片等策略提升性能
5. **安全性**: 网络隔离和权限控制

通过分阶段实施，可以平滑地完成从单体到分布式的转换，确保系统的稳定性和可用性。