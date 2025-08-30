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
	Forward    chan []byte  // 客户端广播通道
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
	IsForward   bool
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
	// 上报给房子
	w.WorkerHouse.AddClient(client.UUID, w)
}

// RemoveClient 从该工作者管理的client列表中移除一个client
func (w *Worker) RemoveClient(client *Client) {
	w.Clients.Delete(client.UUID)
	// 从房子里移除
	w.WorkerHouse.RemoveClient(client.UUID)
}

func (s *Worker) BotSendMessage(to string, content string) {
	// 是机器人助手， 将消息发给DeepSeek
	deepSeekResp, err := s.DeepSeekClient.ChatCompletion(context.Background(), []Message{
		{
			Role:    "user",
			Content: content,
		},
	})
	if err != nil {
		logger.Error("DeepSeek聊天失败", zap.Error(err))
		return
	}
	botMsg := &protocol.Message{
		From:         s.BotClient.UUID,
		FromUsername: s.BotClient.Username,
		To:           to,
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
	client := s.GetClient(to)
	// 回复用户，也就是机器人助手向用户发送消息
	s.SendMessageToClient(client, botMsgByte)
}

// 处理单聊消息的统一逻辑
func (s *Worker) handleSingleMessage(msg *protocol.Message, originalMessage []byte) {
	// 序列化消息
	msgByte, err := proto.Marshal(msg)
	if err != nil {
		logger.Error("消息序列化失败", zap.Error(err))
		return
	}
	// 检查是否是机器人助手
	if msg.To == s.BotClient.UUID {
		// 是机器人助手， 将消息发给DeepSeek
		s.BotSendMessage(msg.From, msg.Content)
	} else {
		isOnline := s.isUserOnline(msg.To)
		if isOnline {
			// 尝试从本地客户端列表中查找目标客户端
			if client, ok := s.Clients.Load(msg.To); ok {
				// 优先使用Kafka发送，否则使用WebSocket
				if dao.KafkaProducerInstance != nil {
					if err := dao.KafkaProducerInstance.SendChatMessage(msg.To, msgByte, s.ID); err != nil {
						logger.Error("发送消息到Kafka失败", zap.Error(err))
						s.SendMessageToClient(client.(*Client), msgByte)
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
		}
	}
}

// 处理群聊消息
func (s *Worker) handleGroupMessage(msg *protocol.Message, originalMessage []byte) {
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

// 处理转发过来的消息
func (s *Worker) handleForwardMessage(msg *protocol.Message) {
	fmt.Println("开始处理转发的消息", msg)
	// 序列化消息
	msgByte, err := proto.Marshal(msg)
	if err != nil {
		logger.Error("消息序列化失败", zap.Error(err))
		return
	}
	isOnline := s.isUserOnline(msg.To)
	if isOnline {
		if client, ok := s.Clients.Load(msg.To); ok {
			// 优先使用Kafka发送，否则使用WebSocket
			if dao.KafkaProducerInstance != nil {
				if err := dao.KafkaProducerInstance.SendChatMessage(msg.To, msgByte, s.ID); err != nil {
					logger.Error("发送消息到Kafka失败", zap.Error(err))
					s.SendMessageToClient(client.(*Client), msgByte)
				}
			} else {
				s.SendMessageToClient(client.(*Client), msgByte)
			}
		}
	} else {
		// 用户离线，消息已保存到数据库，等待用户上线时推送
		s.saveMessageToDB(msg)
	}
}

// 工作者做任务
func (s *Worker) Do() {
	// 启动消息处理队列
	s.startMessageProcessors(3)
	for {
		select {
		case conn := <-s.Register:
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
			// s.Clients.Store(conn.UUID, conn)
			s.AddClient(conn)
			// 直接存储到Redis中表示在线
			ctx := context.Background()
			key := fmt.Sprintf("user:online:%s", conn.UUID)
			dao.REDIS.SetBit(ctx, key, 0, 1)
			// 获知服务自己的worker id
			conn.workerID = s.ID
			// 获知自己的bot id
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
			s.queueMessage(message, false)
		case forward := <-s.Forward:
			s.queueMessage(forward, true)
		}
	}
}

// forwardToOtherWorkers 转发消息到其他worker
func (s *Worker) forwardToOtherWorkers(targetUUID string, message []byte) {
	// 从房子处获取该用户的worker
	worker := s.WorkerHouse.FindWorkerByClientUUID(targetUUID)
	if worker != nil {
		fmt.Printf("worker %d 转发给 worker %d，目标用户: %s\n", s.ID, worker.ID, targetUUID)
		select {
		case worker.Forward <- message:
			fmt.Printf("worker %d 成功发送转发消息到 worker %d\n", s.ID, worker.ID)
		default:
			fmt.Printf("worker %d 转发消息到 worker %d 失败：Forward通道已满\n", s.ID, worker.ID)
		}
		return
	} else {
		fmt.Printf("worker %d 找不到目标用户 %s 的worker\n", s.ID, targetUUID)
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
	fmt.Printf("worker %d 开始处理消息\n", s.ID)
	// 转发消息不需要有群组类型，因为目的很准确
	if task.IsForward {
		fmt.Printf("worker %d 开始处理转发消息\n", s.ID)
		s.handleForwardMessage(task.Message)
		return
	} else {
		if task.Message.MessageType == 1 {
			// 单聊消息
			fmt.Printf("worker %d 开始处理单聊消息\n", s.ID)
			s.handleSingleMessage(task.Message, task.RawData)
		} else {
			// 群聊消息
			fmt.Printf("worker %d 开始处理群聊消息\n", s.ID)
			s.handleGroupMessage(task.Message, task.RawData)
		}
		// // 处理消息
		// switch task.Message.ContentType {
		// case 1, 2, 3, 4, 5, 6, 7, 13: // 文本、文件、图片、语音、视频消息

		// case 8: // 好友请求消息
		// 	if task.Message.MessageType == 1 {
		// 		// 好友请求消息
		// 		s.handleSingleMessage(task.Message, task.RawData)
		// 	}
		// default:
		// 	logger.Warn("未知的消息内容类型",
		// 		zap.Int32("contentType", task.Message.ContentType),
		// 		zap.String("from", task.Message.From),
		// 		zap.String("to", task.Message.To))
		// }
	}
}

func (s *Worker) queueMessage(rawMessage []byte, isForward bool) {
	msg := &protocol.Message{}
	if err := proto.Unmarshal(rawMessage, msg); err != nil {
		logger.Error("解析消息失败", zap.Error(err))
		return
	}
	if msg.To == "" {
		return
	}
	if isForward {
		fmt.Printf("worker %d 收到转发消息，从 %s 到 %s，消息ID: %s\n", s.ID, msg.From, msg.To, msg.MessageId)
	} else {
		fmt.Printf("worker %d 收到非转发消息，从 %s 到 %s，消息ID: %s\n", s.ID, msg.From, msg.To, msg.MessageId)
	}
	task := &MessageTask{
		Message:     msg,
		RawData:     rawMessage,
		ProcessTime: time.Now(),
		IsForward:   isForward,
	}
	select {
	case s.MessageQueue <- task:
		if isForward {
			fmt.Printf("worker %d 转发消息成功入队，消息ID: %s\n", s.ID, msg.MessageId)
		}
	default:
		logger.Warn("消息队列已满，丢弃消息")
		fmt.Printf("worker %d 消息队列已满，丢弃消息ID: %s\n", s.ID, msg.MessageId)
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
	s.RemoveClient(client)
	// 发送Kafka下线事件
	if dao.KafkaProducerInstance != nil {
		// ... Kafka事件发送逻辑
		if err := dao.KafkaProducerInstance.SendUserEvent("user_offline", client.UUID, s.ID, nil); err != nil {
			logger.Error("发送用户下线事件到Kafka失败", zap.Error(err))
		}
	}
	s.TaskCount++
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

// fixme 会重复发消息 发送群聊消息
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
	}
	return false
}
