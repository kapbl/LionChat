package service

import (
	"cchat/internal/dao"
	"cchat/internal/dao/model"
	"cchat/pkg/cgoroutinue"
	"cchat/pkg/logger"
	"cchat/pkg/protocol"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"
)

// 一个worker 可以服务多个 client, 10个client， 后续可以自动扩容
type Worker struct {
	ID         int          // 该工作者的ID
	Clients    sync.Map     // 该工作者管理的client
	Register   chan *Client // 客户端注册通道
	Unregister chan *Client // 客户端注销通道
	Broadcast  chan []byte  // 客户端广播通道
	mutex      sync.RWMutex // 读写锁

	FragmentManager *FragmentManager  // 消息分片管理器
	TaskCount       int               // 该工作者当前管理的任务数量
	WorkerHouse     *WorkerHouse      // 该工作者所在的房子
	MessageQueue    chan *MessageTask // 消息处理队列

	BotClient      *BotClient      // 该工作者管理的机器人
	DeepSeekClient *DeepSeekClient // DeepSeek API客户端

}

type MessageTask struct {
	Message     *protocol.Message
	RawData     []byte
	ProcessTime time.Time
}

// 启动多个消息处理协程
func (s *Worker) startMessageProcessors(count int) {
	for range count {
		// 从goroutine池获取一个goroutine
		cgoroutinue.GoroutinePool.Submit(func() {
			s.messageProcessor()
		})
	}
}

// 消息处理goroutine
func (s *Worker) messageProcessor() {
	// 循环一直执行
	for task := range s.MessageQueue {
		s.processMessageTask(task)
	}
}

// AddClient 添加一个client到该工作者管理的client列表中
func (w *Worker) AddClient(client *Client) {
	w.Clients.Store(client.UUID, client)
}

// RemoveClient 从该工作者管理的client列表中移除一个client
func (w *Worker) RemoveClient(client *Client) {
	w.Clients.Delete(client.UUID)
}

// getContentTypeName 获取内容类型名称
func (s *Worker) getContentTypeName(contentType int32) string {
	switch contentType {
	case 1:
		return "文本"
	case 2:
		return "文件"
	case 3:
		return "图片"
	case 4:
		return "语音"
	case 5:
		return "视频"
	case 6:
		return "语音通话"
	case 7:
		return "视频电话"
	case 8:
		return "好友请求"
	default:
		return "未知"
	}
}

// handleDirectMessage 处理单聊消息的统一逻辑
func (s *Worker) handleDirectMessage(msg *protocol.Message, originalMessage []byte) {
	// 序列化消息
	msgByte, err := proto.Marshal(msg)
	if err != nil {
		logger.Error("消息序列化失败", zap.Error(err))
		return
	}
	// 检查是否是机器人助手
	if msg.To == s.BotClient.UUID {
		// 是机器人助手， 将消息发给DeepSeek
		deepSeekResp, err := s.DeepSeekClient.ChatCompletion(context.Background(), []Message{
			{
				Role:    "user",
				Content: msg.Content,
			},
		})
		if err != nil {
			logger.Error("DeepSeek聊天失败", zap.Error(err))
			return
		}
		logger.Info("DeepSeek聊天成功", zap.Any("response", deepSeekResp))
		// s.handleBotMessage(msg)
		botMsg := &protocol.Message{
			From:         s.BotClient.UUID,
			FromUsername: s.BotClient.Username,
			To:           msg.From,
			MessageId:    strconv.FormatInt(time.Now().UnixNano(), 10),
			ContentType:  1,
			Content:      deepSeekResp.Choices[0].Message.Content,
		}
		// 序列化消息
		botMsgByte, err := proto.Marshal(botMsg)
		if err != nil {
			logger.Error("消息序列化失败", zap.Error(err))
			return
		}
		client := s.GetClient(msg.From)
		// 回复用户，也就是机器人助手向用户发送消息
		s.SendMessageToClient(client, botMsgByte)
		return
	}

	// 检查目标用户是否在线
	isOnline := s.isUserOnline(msg.To)
	if isOnline {
		// 尝试从本地客户端列表中查找目标客户端
		if client, ok := s.Clients.Load(msg.To); ok {
			// 找到本地客户端，直接发送
			logger.Info("发送单聊消息",
				zap.String("type", s.getContentTypeName(msg.ContentType)),
				zap.String("to", msg.To),
				zap.Int("workerID", s.ID))

			// 优先使用Kafka发送，否则使用WebSocket
			if dao.KafkaProducerInstance != nil {
				if err := dao.KafkaProducerInstance.SendChatMessage(msg.To, msgByte, s.ID); err != nil {
					logger.Error("发送消息到Kafka失败", zap.Error(err))
					// Kafka失败时降级到WebSocket
					s.SendMessageToClient(client.(*Client), msgByte)
				} else {
					// Kafka发送成功，不需要再通过WebSocket发送
					logger.Debug("消息已通过Kafka发送", zap.String("to", msg.To))
				}
			} else {
				s.SendMessageToClient(client.(*Client), msgByte)
			}
		} else {
			// 本地未找到，转发到其他worker
			s.forwardToOtherWorkers(msg.To, originalMessage)
		}
	} else {
		// 用户离线，消息已保存到数据库，等待用户上线时推送
		s.saveMessageToDB(msg)
		logger.Info("用户离线，消息已保存到数据库",
			zap.String("to", msg.To),
			zap.String("from", msg.From),
			zap.String("messageId", msg.MessageId))
	}
}

// forwardToOtherWorkers 转发消息到其他worker
func (s *Worker) forwardToOtherWorkers(targetUUID string, message []byte) {
	for _, worker := range s.WorkerHouse.Workers {
		if worker.ID != s.ID {
			if _, ok := worker.Clients.Load(targetUUID); ok {
				logger.Debug("转发消息到其他worker",
					zap.String("target", targetUUID),
					zap.Int("fromWorker", s.ID),
					zap.Int("toWorker", worker.ID))
				worker.Broadcast <- message
				return
			}
		}
	}
	logger.Warn("未找到目标客户端", zap.String("target", targetUUID))
}

// handleGroupMessage 处理群聊消息
func (s *Worker) handleGroupMessage(msg *protocol.Message, originalMessage []byte) {
	logger.Info("处理群聊消息",
		zap.String("type", s.getContentTypeName(msg.ContentType)),
		zap.String("groupID", msg.To))
	s.SendGroupMessage(msg.From, msg.To, originalMessage)
	// 优先使用Kafka发送，否则使用WebSocket
	if dao.KafkaProducerInstance != nil {
		if err := dao.KafkaProducerInstance.SendGroupMessage(msg.To, msg.From, originalMessage); err != nil {
			logger.Error("发送消息到Kafka失败", zap.Error(err))
			// Kafka失败时降级到WebSocket
			s.SendGroupMessage(msg.To, msg.From, originalMessage)
		}
	} else {
		s.SendGroupMessage(msg.To, msg.From, originalMessage)
	}
}

// 工作者做任务
func (s *Worker) Do() {
	// 启动消息处理队列
	s.startMessageProcessors(3)
	for {
		select {
		case conn := <-s.Register:
			logger.Info("注册连接", zap.String("uuid", conn.UUID), zap.Int("workerID", s.ID))
			s.Clients.Store(conn.UUID, conn)
			cgoroutinue.GoroutinePool.Submit(func() {
				// 发送用户上线事件到Kafka
				if dao.KafkaProducerInstance != nil {
					metadata := map[string]interface{}{
						"connection_time": conn.ConnTime,
						"client_ip":       conn.RemoteAddr,
						"worker_id":       s.ID,
					}
					if err := dao.KafkaProducerInstance.SendUserEvent("user_online", conn.UUID, s.ID, metadata); err != nil {
						logger.Error("发送用户上线事件到Kafka失败", zap.Error(err))
					}
				}
			})
			// 直接存储到Redis中表示在线
			ctx := context.Background()
			key := fmt.Sprintf("user:online:%s", conn.UUID)
			dao.REDIS.SetBit(ctx, key, 0, 1)

			// 异步更新数据库用户的在线状态并推送离线消息
			// go func() {
			// 	// 更新用户的在线状态为1
			// 	err := dao.DB.Table("users").Where("uuid = ?", conn.UUID).Update("status", 1).Error
			// 	if err != nil {
			// 		logger.Error("更新用户在线状态失败", zap.Error(err))
			// 	}

			// 	// 推送离线消息
			// 	s.pushOfflineMessages(conn)
			// }()

			// 获知服务自己的worker id
			conn.workerID = s.ID
			// 获知自己的bot id
			// key : user_uuid:bot
			// value : bot uuid
			key = fmt.Sprintf("user:%s:bot", conn.UUID)
			botID := s.BotClient.UUID
			dao.REDIS.Set(ctx, key, botID, 0)
			// 启动客户端的读取和写入 goroutine
			cgoroutinue.GoroutinePool.Submit(func() {
				conn.Read()
			})
			cgoroutinue.GoroutinePool.Submit(func() {
				conn.Write()
			})
		case conn := <-s.Unregister:
			s.handleClientDisconnect(conn)
		case message := <-s.Broadcast:
			s.queueMessage(message)
		}
	}
}
func (s *Worker) processMessageTask(task *MessageTask) {
	// 处理消息任务
	defer func() {
		if r := recover(); r != nil {
			logger.Error("消息处理发生panic",
				zap.Any("panic", r),
				zap.String("messageId", task.Message.MessageId),
				zap.String("from", task.Message.From),
				zap.String("to", task.Message.To))
		}
	}()
	// 处理消息
	switch task.Message.ContentType {
	case 1, 2, 3, 4, 5, 6, 7, 13: // 文本、文件、图片、语音、视频消息
		if task.Message.MessageType == 1 {
			// 单聊消息
			s.handleDirectMessage(task.Message, task.RawData)
		} else {
			// 群聊消息
			s.handleGroupMessage(task.Message, task.RawData)
		}
	case 8: // 好友请求消息
		if task.Message.MessageType == 1 {
			// 好友请求消息
			s.handleDirectMessage(task.Message, task.RawData)
		}
	default:
		logger.Warn("未知的消息内容类型",
			zap.Int32("contentType", task.Message.ContentType),
			zap.String("from", task.Message.From),
			zap.String("to", task.Message.To))
	}
}

func (s *Worker) queueMessage(rawMessage []byte) {
	msg := &protocol.Message{}
	if err := proto.Unmarshal(rawMessage, msg); err != nil {
		logger.Error("解析消息失败", zap.Error(err))
		return
	}
	if msg.To == "" {
		return
	}
	task := &MessageTask{
		Message:     msg,
		RawData:     rawMessage,
		ProcessTime: time.Now(),
	}
	select {
	case s.MessageQueue <- task:
		// 成功入队
	default:
		logger.Warn("消息队列已满，丢弃消息")
	}
}

func (s *Worker) handleClientDisconnect(client *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查客户端是否已经被清理
	if _, exists := s.Clients.Load(client.UUID); !exists {
		return // 已经被清理过了
	}
	// 安全关闭发送通道
	select {
	case <-client.Send:
		// 通道已关闭
	default:
		close(client.Send)
	}
	s.Clients.Delete(client.UUID)
	// 发送Kafka下线事件
	if dao.KafkaProducerInstance != nil {
		// ... Kafka事件发送逻辑
		if err := dao.KafkaProducerInstance.SendUserEvent("user_offline", client.UUID, s.ID, nil); err != nil {
			logger.Error("发送用户下线事件到Kafka失败", zap.Error(err))
		}
	}
	s.TaskCount++
	// 异步更新数据库用户的在线状态
	// go func() {
	// 	// 更新用户的在线状态为0
	// 	err := dao.DB.Table("users").Where("uuid = ?", client.UUID).Update("status", 0).Error
	// 	if err != nil {
	// 		logger.Error("更新用户在线状态失败", zap.Error(err))
	// 	}
	// }()
	ctx := context.Background()
	key := fmt.Sprintf("user:online:%s", client.UUID)
	dao.REDIS.SetBit(ctx, key, 0, 0)

	key = fmt.Sprintf("user:%s:bot", client.UUID)
	dao.REDIS.Del(ctx, key)

	logger.Info("客户端连接已清理", zap.String("uuid", client.UUID))
}

// 使用 uuid 获取客户端
func (s *Worker) GetClient(uuid string) *Client {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	client, ok := s.Clients.Load(uuid)
	if ok {
		return client.(*Client)
	}
	return nil
}

// 发送单个消息（支持分片）
func (s *Worker) SendMessageToClient(client *Client, msg []byte) {
	// 反序列化消息以检查是否需要分片
	var message protocol.Message
	if err := proto.Unmarshal(msg, &message); err != nil {
		logger.Error("反序列化消息失败", zap.Error(err))
		return
	}

	// 检查是否需要分片
	if s.FragmentManager.ShouldFragment(&message) {
		logger.Info("消息需要分片",
			zap.String("to", client.UUID),
			zap.Int("messageSize", len(msg)))

		// 分片消息
		fragments, err := s.FragmentManager.FragmentMessage(&message)
		if err != nil {
			logger.Error("消息分片失败", zap.Error(err))
			return
		}

		// 发送所有分片
		for _, fragment := range fragments {
			fragmentBytes, err := proto.Marshal(fragment)
			if err != nil {
				logger.Error("分片序列化失败", zap.Error(err))
				return
			}
			s.sendRawMessage(client, fragmentBytes)
		}
	} else {
		// 直接发送非分片消息
		s.sendRawMessage(client, msg)
	}
}

func (s *Worker) SendBotMessage(bot *BotClient, msg []byte) {
	bot.Send <- msg
}

// 发送原始消息字节
func (s *Worker) sendRawMessage(client *Client, msg []byte) {
	client.Send <- msg
}

// 发送群聊消息
func (s *Worker) SendGroupMessage(fromUUID string, groupUUID string, msg []byte) {
	// 获取该群聊下的所有群成员的UUID
	s.mutex.Lock()
	defer s.mutex.Unlock()
	groupMembers, err := GetGroupMember(groupUUID)
	if err != nil {
		logger.Error("获取群成员失败", zap.Error(err), zap.String("groupId", groupUUID))
		return
	}
	for _, clientID := range groupMembers {
		if clientID == fromUUID {
			continue
		}
		// 尝试从本地客户端列表中查找目标客户端
		if client, ok := s.Clients.Load(clientID); ok {
			// 找到本地客户端，直接发送
			logger.Info("发送群聊消息",
				zap.String("to", clientID),
				zap.Int("workerID", s.ID))
			s.SendMessageToClient(client.(*Client), msg)
		} else {
			// 本地未找到，转发到其他worker
			s.forwardToOtherWorkers(clientID, msg)
		}
	}
}

// saveMessageToDB 保存消息到数据库
func (s *Worker) saveMessageToDB(msg *protocol.Message) {
	// 使用消息ID（如果没有的话生成一个）
	var messageID string
	if msg.MessageId != "" {
		messageID = msg.MessageId
	} else {
		// 生成一个基于时间戳的唯一ID
		messageID = strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	// 创建消息记录
	message := &model.Message{
		SenderID:  msg.From,
		ReceiveID: msg.To,
		Content:   msg.Content,
		Status:    0, // 0表示未读
		MessageID: messageID,
	}

	// 保存到数据库
	if err := dao.DB.Create(message).Error; err != nil {
		logger.Error("保存消息到数据库失败",
			zap.Error(err),
			zap.String("from", msg.From),
			zap.String("to", msg.To),
			zap.String("messageId", msg.MessageId))
	} else {
		logger.Info("消息已保存到数据库",
			zap.String("from", msg.From),
			zap.String("to", msg.To),
			zap.String("dbMessageId", messageID))
	}
}

// isUserOnline 检查用户是否在线
func (s *Worker) isUserOnline(userUUID string) bool {
	// 检查Redis中是否存在
	ctx := context.Background()
	key := fmt.Sprintf("user:online:%s", userUUID)
	online := dao.REDIS.GetBit(ctx, key, 0).Val()
	if online == 1 {
		return true
	} else {
		return false
	}
	// 首先检查本地客户端列表
	if _, ok := s.Clients.Load(userUUID); ok {
		return true
	}

	// 检查其他worker的客户端列表
	for _, worker := range s.WorkerHouse.Workers {
		if worker.ID != s.ID {
			if _, ok := worker.Clients.Load(userUUID); ok {
				return true
			}
		}
	}

	// 最后检查数据库中的用户状态
	var user model.Users
	err := dao.DB.Table("users").Where("uuid = ?", userUUID).First(&user).Error
	if err != nil {
		logger.Error("查询用户状态失败", zap.Error(err), zap.String("uuid", userUUID))
		return false
	}

	return user.Status == 1
}

// pushOfflineMessages 推送离线消息给刚上线的用户
func (s *Worker) pushOfflineMessages(client *Client) {
	// 查询该用户的未读消息
	var messages []model.Message
	err := dao.DB.Table("message").Where("receive_id = ? AND status = 0", client.UUID).Order("created_at ASC").Find(&messages).Error
	if err != nil {
		logger.Error("查询离线消息失败", zap.Error(err), zap.String("uuid", client.UUID))
		return
	}

	if len(messages) == 0 {
		logger.Info("用户无离线消息", zap.String("uuid", client.UUID))
		return
	}

	logger.Info("开始推送离线消息",
		zap.String("uuid", client.UUID),
		zap.Int("count", len(messages)))

	// 逐条推送离线消息
	for _, msg := range messages {
		// 构造协议消息
		protocolMsg := &protocol.Message{
			From:        msg.SenderID,
			To:          msg.ReceiveID,
			Content:     msg.Content,
			ContentType: 1, // 默认为文本消息
			// Type:        1, // 单聊消息
			MessageType: 1, // 普通消息
			MessageId:   msg.MessageID,
			Timestamp:   time.Now().Unix(),
		}

		// 序列化消息
		msgByte, err := proto.Marshal(protocolMsg)
		if err != nil {
			logger.Error("序列化离线消息失败", zap.Error(err), zap.String("messageId", msg.MessageID))
			continue
		}

		// 发送消息给客户端
		s.SendMessageToClient(client, msgByte)

		// 标记消息为已读（可选，也可以等客户端确认后再标记）
		// 这里先不自动标记为已读，等待客户端确认
		logger.Debug("推送离线消息",
			zap.String("from", msg.SenderID),
			zap.String("to", msg.ReceiveID),
			zap.String("messageId", msg.MessageID))
	}

	logger.Info("离线消息推送完成",
		zap.String("uuid", client.UUID),
		zap.Int("count", len(messages)))
}

// handleBotMessage 处理机器人消息，调用DeepSeek API
func (s *Worker) handleBotMessage(msg *protocol.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 检查DeepSeek客户端是否已初始化
	if s.DeepSeekClient == nil {
		logger.Error("DeepSeek客户端未初始化")
		return
	}

	// 构建消息历史（这里简化处理，实际应该维护对话上下文）
	messages := []Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant.",
		},
		{
			Role:    "user",
			Content: msg.Content,
		},
	}

	// 调用DeepSeek API
	response, err := s.DeepSeekClient.ChatCompletion(ctx, messages)
	if err != nil {
		logger.Error("调用DeepSeek API失败", zap.Error(err))
		// 发送错误消息给用户
		s.sendErrorMessageToUser(msg.From, "抱歉，AI助手暂时无法回复，请稍后再试。")
		return
	}

	// 检查响应是否有效
	if len(response.Choices) == 0 {
		logger.Error("DeepSeek API返回空响应")
		s.sendErrorMessageToUser(msg.From, "抱歉，AI助手没有返回有效回复。")
		return
	}

	// 构建回复消息
	botReply := &protocol.Message{
		From:        s.BotClient.UUID,
		To:          msg.From,
		Content:     response.Choices[0].Message.Content,
		ContentType: msg.ContentType,
		Timestamp:   time.Now().Unix(),
		Type:        "text",
		MessageType: 1,
		MessageId:   fmt.Sprintf("bot_%d", time.Now().UnixNano()),
	}

	// 序列化回复消息
	replyByte, err := proto.Marshal(botReply)
	if err != nil {
		logger.Error("机器人回复消息序列化失败", zap.Error(err))
		return
	}

	// 保存消息到数据库
	s.saveMessageToDB(botReply)

	// 发送回复给用户
	if client, ok := s.Clients.Load(msg.From); ok {
		s.SendMessageToClient(client.(*Client), replyByte)
		logger.Info("AI助手回复发送成功", zap.String("to", msg.From))
	} else {
		// 用户不在线，转发到其他worker或保存为离线消息
		s.forwardToOtherWorkers(msg.From, replyByte)
	}
}

// sendErrorMessageToUser 发送错误消息给用户
func (s *Worker) sendErrorMessageToUser(userUUID, errorMsg string) {
	errorMessage := &protocol.Message{
		From:        s.BotClient.UUID,
		To:          userUUID,
		Content:     errorMsg,
		ContentType: 1, // 文本消息
		Timestamp:   time.Now().Unix(),
		Type:        "text",
		MessageType: 1,
		MessageId:   fmt.Sprintf("bot_error_%d", time.Now().UnixNano()),
	}

	errorByte, err := proto.Marshal(errorMessage)
	if err != nil {
		logger.Error("错误消息序列化失败", zap.Error(err))
		return
	}

	if client, ok := s.Clients.Load(userUUID); ok {
		s.SendMessageToClient(client.(*Client), errorByte)
	} else {
		s.forwardToOtherWorkers(userUUID, errorByte)
	}
}
