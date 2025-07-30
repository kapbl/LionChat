package service

import (
	"cchat/internal/dao"
	"cchat/pkg/config"
	"cchat/pkg/logger"
	"context"
	"encoding/json"
	"sync"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// KafkaConsumerService Kafka消费者服务
type KafkaConsumerService struct {
	consumerGroup sarama.ConsumerGroup
	config        *config.Config
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// NewKafkaConsumerService 创建Kafka消费者服务
func NewKafkaConsumerService(cfg *config.Config) (*KafkaConsumerService, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true

	consumerGroup, err := sarama.NewConsumerGroup(cfg.Kafka.Brokers, cfg.Kafka.Consumer.GroupID, config)
	if err != nil {
		logger.Error("创建Kafka消费者组失败", zap.Error(err))
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &KafkaConsumerService{
		consumerGroup: consumerGroup,
		config:        cfg,
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

// Start 启动消费者服务
func (kcs *KafkaConsumerService) Start() {
	topics := []string{
		kcs.config.Kafka.Topics.ChatMessages,
		kcs.config.Kafka.Topics.UserEvents,
		kcs.config.Kafka.Topics.GroupMessages,
	}

	// 启动消费者
	kcs.wg.Add(1)
	go func() {
		defer kcs.wg.Done()
		for {
			select {
			case <-kcs.ctx.Done():
				logger.Info("Kafka消费者服务停止")
				return
			default:
				if err := kcs.consumerGroup.Consume(kcs.ctx, topics, kcs); err != nil {
					logger.Error("Kafka消费失败", zap.Error(err))
				}
			}
		}
	}()

	// 处理错误
	kcs.wg.Add(1)
	go func() {
		defer kcs.wg.Done()
		for {
			select {
			case <-kcs.ctx.Done():
				return
			case err := <-kcs.consumerGroup.Errors():
				logger.Error("Kafka消费者错误", zap.Error(err))
			}
		}
	}()

	logger.Info("Kafka消费者服务启动成功", zap.Strings("topics", topics))
}

// Stop 停止消费者服务
func (kcs *KafkaConsumerService) Stop() {
	kcs.cancel()
	kcs.wg.Wait()
	if err := kcs.consumerGroup.Close(); err != nil {
		logger.Error("关闭Kafka消费者组失败", zap.Error(err))
	}
	logger.Info("Kafka消费者服务已停止")
}

// Setup 实现sarama.ConsumerGroupHandler接口
func (kcs *KafkaConsumerService) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup 实现sarama.ConsumerGroupHandler接口
func (kcs *KafkaConsumerService) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 实现sarama.ConsumerGroupHandler接口
func (kcs *KafkaConsumerService) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			// 处理消息
			if err := kcs.handleMessage(message); err != nil {
				logger.Error("处理Kafka消息失败",
					zap.Error(err),
					zap.String("topic", message.Topic),
					zap.String("key", string(message.Key)))
			} else {
				// 标记消息已处理
				session.MarkMessage(message, "")
			}

		case <-session.Context().Done():
			return nil
		}
	}
}

// handleMessage 处理Kafka消息
func (kcs *KafkaConsumerService) handleMessage(message *sarama.ConsumerMessage) error {
	var event dao.MessageEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		logger.Error("反序列化消息失败", zap.Error(err))
		return err
	}

	logger.Debug("收到Kafka消息",
		zap.String("topic", message.Topic),
		zap.String("event_type", event.EventType),
		zap.String("user_id", event.UserID),
		zap.Int64("timestamp", event.Timestamp))

	// 根据主题和事件类型处理消息
	switch message.Topic {
	case kcs.config.Kafka.Topics.ChatMessages:
		return kcs.handleChatMessage(&event)
	case kcs.config.Kafka.Topics.UserEvents:
		return kcs.handleUserEvent(&event)
	case kcs.config.Kafka.Topics.GroupMessages:
		return kcs.handleGroupMessage(&event)
	default:
		logger.Warn("未知的Kafka主题", zap.String("topic", message.Topic))
	}

	return nil
}

// handleChatMessage 处理聊天消息
func (kcs *KafkaConsumerService) handleChatMessage(event *dao.MessageEvent) error {
	logger.Info("处理聊天消息事件",
		zap.String("user_id", event.UserID),
		zap.String("event_type", event.EventType))

	// 将消息发送给目标用户
	if ServerInstance != nil {
		// receiverID, ok := event.Metadata["receiver_id"].(string)
		// if !ok {
		// 	logger.Warn("聊天消息缺少 receiver_id")
		// 	return nil
		// }
		client := ServerInstance.GetClient(event.UserID)
		if client != nil {
			// 转换为 []byte 类型
			messageData, ok := event.MessageData.([]byte)
			if !ok {
				logger.Warn("聊天消息数据类型错误")
				return nil
			}
			ServerInstance.SendMessageToClient(client, messageData)
		}
	}

	return nil
}

// handleUserEvent 处理用户事件
func (kcs *KafkaConsumerService) handleUserEvent(event *dao.MessageEvent) error {
	logger.Info("处理用户事件",
		zap.String("user_id", event.UserID),
		zap.String("event_type", event.EventType))

	// 根据事件类型处理
	switch event.EventType {
	case "user_online":
		// 处理用户上线事件
		logger.Info("用户上线", zap.String("user_id", event.UserID))
	case "user_offline":
		// 处理用户下线事件
		logger.Info("用户下线", zap.String("user_id", event.UserID))
	case "user_typing":
		// 处理用户正在输入事件
		logger.Debug("用户正在输入", zap.String("user_id", event.UserID))
	}

	return nil
}

// handleGroupMessage 处理群组消息
func (kcs *KafkaConsumerService) handleGroupMessage(event *dao.MessageEvent) error {
	groupID, ok := event.Metadata["group_id"].(string)
	if !ok {
		logger.Warn("群组消息缺少 group_id")
		return nil
	}

	logger.Info("处理群组消息事件",
		zap.String("group_id", groupID),
		zap.String("event_type", event.EventType))
	// 这里可以添加群组消息相关的处理逻辑
	// 例如：消息持久化、成员通知等

	return nil
}
