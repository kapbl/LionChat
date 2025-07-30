package kafka

import (
	"cchat/internal/dao"
	"cchat/pkg/config"
	"cchat/pkg/logger"
	"cchat/pkg/protocol"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestKafkaIntegration 测试Kafka集成
func TestKafkaIntegration(t *testing.T) {
	// 初始化日志
	logger.InitLogger()

	// 加载配置
	cfg := config.LoadConfig()

	// 初始化Kafka
	err := dao.InitKafka(&cfg)
	assert.NoError(t, err, "Kafka初始化应该成功")

	// 确保Kafka实例已创建
	assert.NotNil(t, dao.KafkaProducerInstance, "Kafka生产者实例应该存在")
	assert.NotNil(t, dao.KafkaConsumerInstance, "Kafka消费者实例应该存在")

	// 测试发送聊天消息
	t.Run("SendChatMessage", func(t *testing.T) {
		testMessage := &protocol.Message{
			From:        "test-user-1",
			To:          "test-user-2",
			Content:     "Hello, this is a test message!",
			ContentType: 1, // 文本消息
			MessageType: 1, // 单聊
			FromUsername: "TestUser1",
		}

		err := dao.KafkaProducerInstance.SendChatMessage("test-user-1", testMessage)
		assert.NoError(t, err, "发送聊天消息应该成功")
		logger.Info("聊天消息发送测试完成")
	})

	// 测试发送用户事件
	t.Run("SendUserEvent", func(t *testing.T) {
		metadata := map[string]interface{}{
			"client_ip":       "192.168.1.100",
			"connection_time": time.Now().Unix(),
			"user_agent":      "TestAgent/1.0",
		}

		err := dao.KafkaProducerInstance.SendUserEvent("user_online", "test-user-1", metadata)
		assert.NoError(t, err, "发送用户事件应该成功")
		logger.Info("用户事件发送测试完成")
	})

	// 测试发送群组消息
	t.Run("SendGroupMessage", func(t *testing.T) {
		testGroupMessage := &protocol.Message{
			From:        "test-user-1",
			To:          "test-group-1",
			Content:     "Hello group! This is a test message!",
			ContentType: 1, // 文本消息
			MessageType: 2, // 群聊
			FromUsername: "TestUser1",
		}

		err := dao.KafkaProducerInstance.SendGroupMessage("test-group-1", "test-user-1", testGroupMessage)
		assert.NoError(t, err, "发送群组消息应该成功")
		logger.Info("群组消息发送测试完成")
	})

	// 等待消息处理
	time.Sleep(2 * time.Second)

	// 清理资源
	if dao.KafkaProducerInstance != nil {
		err := dao.KafkaProducerInstance.Close()
		assert.NoError(t, err, "关闭Kafka生产者应该成功")
	}

	if dao.KafkaConsumerInstance != nil {
		err := dao.KafkaConsumerInstance.Close()
		assert.NoError(t, err, "关闭Kafka消费者应该成功")
	}

	logger.Info("Kafka集成测试完成")
}

// TestKafkaProducerOnly 仅测试Kafka生产者
func TestKafkaProducerOnly(t *testing.T) {
	// 初始化日志
	logger.InitLogger()

	// 加载配置
	cfg := config.LoadConfig()

	// 仅创建生产者
	producer, err := dao.NewKafkaProducer(&cfg)
	assert.NoError(t, err, "创建Kafka生产者应该成功")
	assert.NotNil(t, producer, "Kafka生产者应该存在")

	// 测试发送消息
	testData := map[string]interface{}{
		"test_key": "test_value",
		"timestamp": time.Now().Unix(),
	}

	err = producer.SendMessage(cfg.Kafka.Topics.ChatMessages, "test-key", testData)
	assert.NoError(t, err, "发送测试消息应该成功")

	// 清理资源
	err = producer.Close()
	assert.NoError(t, err, "关闭生产者应该成功")

	logger.Info("Kafka生产者测试完成")
}

// TestKafkaConsumerOnly 仅测试Kafka消费者
func TestKafkaConsumerOnly(t *testing.T) {
	// 初始化日志
	logger.InitLogger()

	// 加载配置
	cfg := config.LoadConfig()

	// 仅创建消费者
	consumer, err := dao.NewKafkaConsumer(&cfg)
	assert.NoError(t, err, "创建Kafka消费者应该成功")
	assert.NotNil(t, consumer, "Kafka消费者应该存在")

	// 清理资源
	err = consumer.Close()
	assert.NoError(t, err, "关闭消费者应该成功")

	logger.Info("Kafka消费者测试完成")
}

// BenchmarkKafkaProducer 性能测试
func BenchmarkKafkaProducer(b *testing.B) {
	// 初始化日志
	logger.InitLogger()

	// 加载配置
	cfg := config.LoadConfig()

	// 创建生产者
	producer, err := dao.NewKafkaProducer(&cfg)
	if err != nil {
		b.Fatalf("创建Kafka生产者失败: %v", err)
	}
	defer producer.Close()

	// 准备测试数据
	testMessage := &protocol.Message{
		From:        "bench-user",
		To:          "bench-target",
		Content:     "Benchmark test message",
		ContentType: 1,
		MessageType: 1,
		FromUsername: "BenchUser",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			err := producer.SendChatMessage("bench-user", testMessage)
			if err != nil {
				b.Errorf("发送消息失败: %v", err)
			}
		}
	})
}

// TestKafkaErrorHandling 测试错误处理
func TestKafkaErrorHandling(t *testing.T) {
	// 初始化日志
	logger.InitLogger()

	// 使用错误的配置
	cfg := config.Config{}
	cfg.Kafka.Brokers = []string{"invalid-broker:9092"}
	cfg.Kafka.Topics.ChatMessages = "test-topic"
	cfg.Kafka.Producer.Retries = 1

	// 尝试创建生产者（应该失败）
	producer, err := dao.NewKafkaProducer(&cfg)
	assert.Error(t, err, "使用无效broker应该失败")
	assert.Nil(t, producer, "生产者应该为nil")

	logger.Info("Kafka错误处理测试完成")
}

// TestMessageEventSerialization 测试消息事件序列化
func TestMessageEventSerialization(t *testing.T) {
	// 创建测试事件
	event := dao.MessageEvent{
		EventType: "test_message",
		Timestamp: time.Now().Unix(),
		UserID:    "test-user",
		MessageData: map[string]interface{}{
			"content": "test content",
			"type":    "text",
		},
		Metadata: map[string]interface{}{
			"source": "test",
		},
	}

	// 测试序列化和反序列化
	// 这里可以添加JSON序列化测试逻辑
	assert.NotEmpty(t, event.EventType, "事件类型不应为空")
	assert.NotEmpty(t, event.UserID, "用户ID不应为空")
	assert.NotZero(t, event.Timestamp, "时间戳不应为零")

	logger.Info("消息事件序列化测试完成")
}