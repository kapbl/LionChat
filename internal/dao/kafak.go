package dao

import (
	"cchat/pkg/config"
	"cchat/pkg/logger"
	"encoding/json"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// KafkaProducer Kafka生产者
type KafkaProducer struct {
	producer sarama.SyncProducer
	config   *config.Config
}

// KafkaConsumer Kafka消费者
type KafkaConsumer struct {
	consumer sarama.Consumer
	config   *config.Config
}

// MessageEvent 消息事件结构
type MessageEvent struct {
	EventType   string                 `json:"event_type"` // message, user_online, user_offline, group_join, group_leave
	Timestamp   int64                  `json:"timestamp"`
	UserID      string                 `json:"user_id"`
	MessageData []byte                 `json:"message_data,omitempty"` // 消息数据，二进制
	Metadata    map[string]interface{} `json:"metadata,omitempty"` 
}

var (
	KafkaProducerInstance *KafkaProducer
	KafkaConsumerInstance *KafkaConsumer
)

// InitKafka 初始化Kafka客户端
func InitKafka(cfg *config.Config) error {
	// 初始化生产者
	producer, err := NewKafkaProducer(cfg)
	if err != nil {
		return err
	}
	KafkaProducerInstance = producer

	// 初始化消费者
	consumer, err := NewKafkaConsumer(cfg)
	if err != nil {
		return err
	}
	KafkaConsumerInstance = consumer

	logger.Info("Kafka客户端初始化成功")
	return nil
}

// NewKafkaProducer 创建Kafka生产者
func NewKafkaProducer(cfg *config.Config) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // 等待所有副本确认
	config.Producer.Retry.Max = cfg.Kafka.Producer.Retries
	config.Producer.Return.Successes = true
	config.Producer.Flush.Frequency = time.Duration(cfg.Kafka.Producer.LingerMs) * time.Millisecond
	config.Producer.Flush.Messages = cfg.Kafka.Producer.BatchSize

	producer, err := sarama.NewSyncProducer(cfg.Kafka.Brokers, config)
	if err != nil {
		logger.Error("创建Kafka生产者失败", zap.Error(err))
		return nil, err
	}

	return &KafkaProducer{
		producer: producer,
		config:   cfg,
	}, nil
}

// NewKafkaConsumer 创建Kafka消费者
func NewKafkaConsumer(cfg *config.Config) (*KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	// 设置消费者组的偏移量重置策略
	switch cfg.Kafka.Consumer.AutoOffsetReset {
	case "earliest":
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	case "latest":
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	default:
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	}

	consumer, err := sarama.NewConsumer(cfg.Kafka.Brokers, config)
	if err != nil {
		logger.Error("创建Kafka消费者失败", zap.Error(err))
		return nil, err
	}

	return &KafkaConsumer{
		consumer: consumer,
		config:   cfg,
	}, nil
}

// SendMessage 发送消息到Kafka
func (kp *KafkaProducer) SendMessage(topic string, key string, message interface{}) error {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		logger.Error("序列化消息失败", zap.Error(err))
		return err
	}
	
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Key:       sarama.StringEncoder(key),
		Value:     sarama.ByteEncoder(msgBytes),
		Timestamp: time.Now(),
	}

	partition, offset, err := kp.producer.SendMessage(msg)
	if err != nil {
		logger.Error("发送消息到Kafka失败",
			zap.Error(err),
			zap.String("topic", topic),
			zap.String("key", key))
		return err
	}

	logger.Debug("消息发送成功",
		zap.String("topic", topic),
		zap.String("key", key),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset))

	return nil
}

// SendChatMessage 发送聊天消息事件
func (kp *KafkaProducer) SendChatMessage(userID string, messageData []byte) error {
	event := MessageEvent{
		EventType:   "message",
		Timestamp:   time.Now().Unix(),
		UserID:      userID,
		MessageData: messageData,
	}

	return kp.SendMessage(kp.config.Kafka.Topics.ChatMessages, userID, event)
}

// SendUserEvent 发送用户事件（上线/下线等）
func (kp *KafkaProducer) SendUserEvent(eventType, userID string, metadata map[string]interface{}) error {
	event := MessageEvent{
		EventType: eventType,
		Timestamp: time.Now().Unix(),
		UserID:    userID,
		Metadata:  metadata,
	}

	return kp.SendMessage(kp.config.Kafka.Topics.UserEvents, userID, event)
}

// SendGroupMessage 发送群组消息事件
func (kp *KafkaProducer) SendGroupMessage(groupID, userID string, messageData []byte) error {
	event := MessageEvent{
		EventType:   "group_message",
		Timestamp:   time.Now().Unix(),
		UserID:      userID,
		MessageData: messageData,
		Metadata: map[string]interface{}{
			"group_id": groupID,
		},
	}

	return kp.SendMessage(kp.config.Kafka.Topics.GroupMessages, groupID, event)
}

// Close 关闭生产者
func (kp *KafkaProducer) Close() error {
	if kp.producer != nil {
		return kp.producer.Close()
	}
	return nil
}

// Close 关闭消费者
func (kc *KafkaConsumer) Close() error {
	if kc.consumer != nil {
		return kc.consumer.Close()
	}
	return nil
}
