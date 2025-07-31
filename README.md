```plaintext
your_project/
├── cmd/                    # 每个可执行程序一个子目录（main 入口）
│   └── yourapp/            # 例如：main.go
│       └── main.go
├── internal/               # 项目私有逻辑（不对外暴露）
│   ├── service/            # 业务逻辑层（如 UserService）
│   ├── handler/            # HTTP handler 层（或 controller）
│   └── dao/                # 数据访问层（如数据库、Redis）
├── pkg/                    # 可被其他项目复用的公共库（类似工具库）
│   └── utils/              # 工具包（如加密、日志封装）
├── api/                    # API 定义（OpenAPI/Swagger, Protobuf 等）
│   └── v1/                 # v1 版本 API 接口定义
├── config/                 # 配置文件（如 yaml/json/toml）
│   └── config.yaml
├── migrations/             # 数据库迁移脚本（可配合 goose, migrate 等工具）
│   └── 001_create_user.sql
├── web/                    # 前端静态资源（如 HTML、JS、CSS）
│   └── static/
├── test/                   # 项目的测试代码
│   └── yourapp_test.go
├── scripts/                # 运维或自动化脚本（如构建、部署）
│   └── build.sh
├── go.mod                  # Go modules 配置文件
├── go.sum
├── .env                    # 环境变量文件（如数据库配置）
├── .gitignore
└── README.md
```

## 🔧 Kafka集成使用方法

### 环境要求

- Go 1.19+
- Apache Kafka 2.8+
- Zookeeper (如果使用Kafka 2.8以下版本)

### 快速开始

1. **启动Kafka服务**
   ```bash
   # 使用Docker Compose启动Kafka
   docker-compose up -d kafka zookeeper
   ```

2. **配置Kafka连接**
   
   在 `config/config.yaml` 中配置Kafka连接信息：
   ```yaml
   kafka:
     brokers:
       - "localhost:9092"
     topics:
       chat_messages: "chat-messages"
       user_events: "user-events"
     consumer_group: "chatlion-group"
   ```

3. **运行应用**
   ```bash
   go run cmd/yourapp/main.go
   ```

### Kafka主题说明

| 主题名称 | 用途 | 消息格式 |
|---------|------|----------|
| `chat-messages` | 聊天消息传递 | JSON格式的消息对象 |
| `user-events` | 用户上线/下线事件 | JSON格式的用户事件 |

### 消息格式示例

**聊天消息格式：**
```json
{
  "id": "msg_123",
  "from_user_id": "user_456",
  "to_user_id": "user_789",
  "content": "Hello, World!",
  "timestamp": "2024-01-01T12:00:00Z",
  "message_type": "text"
}
```

**用户事件格式：**
```json
{
  "user_id": "user_456",
  "event_type": "online",
  "timestamp": "2024-01-01T12:00:00Z",
  "metadata": {
    "ip_address": "192.168.1.100"
  }
}
```

### 部署配置

**生产环境Kafka配置建议：**

```yaml
kafka:
  brokers:
    - "kafka-1:9092"
    - "kafka-2:9092"
    - "kafka-3:9092"
  producer:
    acks: "all"
    retries: 3
    batch_size: 16384
  consumer:
    auto_offset_reset: "earliest"
    enable_auto_commit: false
```

### 监控和运维

- **查看Kafka主题：**
  ```bash
  kafka-topics.sh --bootstrap-server localhost:9092 --list
  ```

- **查看消费者组状态：**
  ```bash
  kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --group chatlion-group
  ```

## 🛠️ 开发指南

### 添加新的消息类型

1. 在 `internal/service/` 中定义新的消息处理逻辑
2. 在 `pkg/kafka/` 中添加对应的生产者/消费者
3. 更新配置文件中的主题配置

### 扩展事件类型

参考 `internal/handler/` 中的事件处理器实现，添加新的事件类型处理逻辑。
