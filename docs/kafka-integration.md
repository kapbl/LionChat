# Kafka集成说明

## 概述

本项目已集成Apache Kafka作为消息队列系统，用于处理聊天消息、用户事件和群组消息的异步处理。Kafka的集成提供了以下功能：

- **消息持久化**: 所有聊天消息都会发送到Kafka进行持久化存储
- **事件驱动架构**: 用户上线/下线、消息发送等事件通过Kafka进行解耦处理
- **水平扩展**: 支持多个消费者实例处理消息，提高系统吞吐量
- **消息可靠性**: 通过Kafka的副本机制确保消息不丢失

## 架构设计

### 主题(Topics)设计

1. **chat-messages**: 存储所有聊天消息
2. **user-events**: 存储用户相关事件（上线、下线、正在输入等）
3. **group-messages**: 存储群组消息和群组相关事件

### 消息格式

所有Kafka消息都使用统一的`MessageEvent`结构：

```json
{
  "event_type": "message|user_online|user_offline|group_message",
  "timestamp": 1640995200,
  "user_id": "user-uuid",
  "message_data": {
    // protobuf消息数据
  },
  "metadata": {
    "group_id": "group-uuid",
    "client_ip": "192.168.1.100"
  }
}
```

## 配置说明

### 开发环境配置 (config.dev.yaml)

```yaml
kafka:
  brokers:
    - "localhost:9092"
  topics:
    chat_messages: "chat-messages"
    user_events: "user-events"
    group_messages: "group-messages"
  consumer:
    group_id: "chat-consumer-group"
    auto_offset_reset: "latest"
  producer:
    acks: "all"
    retries: 3
    batch_size: 16384
    linger_ms: 1
    buffer_memory: 33554432
```

### 生产环境配置 (config.prod.yaml)

```yaml
kafka:
  brokers:
    - "kafka1:9092"
    - "kafka2:9092"
    - "kafka3:9092"
  topics:
    chat_messages: "chat-messages"
    user_events: "user-events"
    group_messages: "group-messages"
  consumer:
    group_id: "chat-consumer-group"
    auto_offset_reset: "latest"
  producer:
    acks: "all"
    retries: 5
    batch_size: 32768
    linger_ms: 5
    buffer_memory: 67108864
```

## 部署指南

### 1. 安装Kafka

#### 使用Docker Compose

创建`docker-compose.kafka.yml`文件：

```yaml
version: '3.8'
services:
  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    ports:
      - "2181:2181"

  kafka:
    image: confluentinc/cp-kafka:latest
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_AUTO_CREATE_TOPICS_ENABLE: true
```

启动Kafka：
```bash
docker-compose -f docker-compose.kafka.yml up -d
```

### 2. 创建主题

```bash
# 进入Kafka容器
docker exec -it <kafka-container-id> bash

# 创建主题
kafka-topics --create --topic chat-messages --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
kafka-topics --create --topic user-events --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
kafka-topics --create --topic group-messages --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
```

### 3. 验证安装

```bash
# 列出所有主题
kafka-topics --list --bootstrap-server localhost:9092

# 查看主题详情
kafka-topics --describe --topic chat-messages --bootstrap-server localhost:9092
```

## 使用示例

### 发送消息到Kafka

```go
// 发送聊天消息
if dao.KafkaProducerInstance != nil {
    err := dao.KafkaProducerInstance.SendChatMessage(userID, messageData)
    if err != nil {
        logger.Error("发送消息到Kafka失败", zap.Error(err))
    }
}

// 发送用户事件
metadata := map[string]interface{}{
    "client_ip": "192.168.1.100",
}
err := dao.KafkaProducerInstance.SendUserEvent("user_online", userID, metadata)

// 发送群组消息
err := dao.KafkaProducerInstance.SendGroupMessage(groupID, userID, messageData)
```

### 消费消息

消费者服务会自动启动并处理Kafka消息。你可以在`internal/service/kafka_consumer.go`中自定义消息处理逻辑：

```go
// 处理聊天消息
func (kcs *KafkaConsumerService) handleChatMessage(event *dao.MessageEvent) error {
    // 自定义处理逻辑
    // 例如：保存到数据库、发送推送通知等
    return nil
}
```

## 监控和调试

### 查看消费者组状态

```bash
kafka-consumer-groups --bootstrap-server localhost:9092 --describe --group chat-consumer-group
```

### 查看主题消息

```bash
# 从最新消息开始消费
kafka-console-consumer --bootstrap-server localhost:9092 --topic chat-messages --from-beginning

# 查看用户事件
kafka-console-consumer --bootstrap-server localhost:9092 --topic user-events --from-beginning
```

## 扩展功能

### 1. 消息持久化

可以扩展消费者服务，将Kafka消息持久化到数据库：

```go
func (kcs *KafkaConsumerService) handleChatMessage(event *dao.MessageEvent) error {
    // 解析消息数据
    var msg protocol.Message
    if err := json.Unmarshal(event.MessageData, &msg); err != nil {
        return err
    }
    
    // 保存到数据库
    dbMessage := &model.Message{
        From:        msg.From,
        To:          msg.To,
        Content:     msg.Content,
        ContentType: msg.ContentType,
        MessageType: msg.MessageType,
        CreatedAt:   time.Unix(event.Timestamp, 0),
    }
    
    return dao.DB.Create(dbMessage).Error
}
```

### 2. 推送通知

集成推送服务，当收到离线用户消息时发送推送通知：

```go
func (kcs *KafkaConsumerService) handleChatMessage(event *dao.MessageEvent) error {
    // 检查用户是否在线
    if !kcs.isUserOnline(event.UserID) {
        // 发送推送通知
        return kcs.sendPushNotification(event)
    }
    return nil
}
```

### 3. 消息统计

收集消息统计信息，用于分析和监控：

```go
func (kcs *KafkaConsumerService) handleChatMessage(event *dao.MessageEvent) error {
    // 更新消息统计
    kcs.updateMessageStats(event)
    return nil
}
```