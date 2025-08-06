package service

import (
	"cchat/internal/dao"
	"cchat/pkg/logger"
	"cchat/pkg/protocol"
	"sync"

	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"
)

// 一个worker 可以服务多个 client, 10个client

type Worker struct {
	ID              int      // 该工作者的ID
	Clients         sync.Map // 该工作者管理的client
	Register        chan *Client
	Unregister      chan *Client
	Broadcast       chan []byte
	mutex           sync.Mutex
	FragmentManager *FragmentManager // 消息分片管理器
	TaskCount       int              // 该工作者当前管理的任务数量
	WorkerHouse     *WorkerHouse     // 该工作者所在的房子

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
			}
		} else {
			s.SendMessageToClient(client.(*Client), msgByte)
		}
	} else {
		// 本地未找到，转发到其他worker
		s.forwardToOtherWorkers(msg.To, originalMessage)
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

// handleGroupMessage 处理群聊消息（预留接口）
func (s *Worker) handleGroupMessage(msg *protocol.Message, originalMessage []byte) {
	// TODO: 实现群聊消息处理逻辑
	logger.Info("处理群聊消息",
		zap.String("type", s.getContentTypeName(msg.ContentType)),
		zap.String("groupID", msg.To))
}

// 工作者做任务
func (s *Worker) Do() {
	for {
		select {
		case conn := <-s.Register:
			s.mutex.Lock()
			logger.Info("注册连接", zap.String("uuid", conn.UUID), zap.Int("workerID", s.ID))
			s.Clients.Store(conn.UUID, conn)
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
			// 启动客户端的读取和写入 goroutine
			conn.workerID = s.ID
			go conn.Read()
			go conn.Write()
			s.mutex.Unlock()

		case conn := <-s.Unregister:
			s.handleClientDisconnect(conn)

		case message := <-s.Broadcast:
			msg := protocol.Message{}
			if err := proto.Unmarshal(message, &msg); err != nil {
				logger.Error("解析消息失败", zap.Error(err))
				continue
			}

			// 检查是否为有效的目标消息
			if msg.To == "" {
				logger.Debug("收到广播消息，跳过处理")
				continue
			}

			s.mutex.Lock()
			// 使用统一的消息处理逻辑
			switch msg.ContentType {
			case 1, 2, 3, 4, 5: // 文本、文件、图片、语音、视频消息
				if msg.MessageType == 1 {
					// 单聊消息
					s.handleDirectMessage(&msg, message)
				} else {
					// 群聊消息
					s.handleGroupMessage(&msg, message)
				}
			case 8: // 好友请求消息
				if msg.MessageType == 1 {
					s.handleDirectMessage(&msg, message)
				}
			default:
				logger.Warn("未知的消息内容类型",
					zap.Int32("contentType", msg.ContentType),
					zap.String("from", msg.From),
					zap.String("to", msg.To))
			}
			s.mutex.Unlock()
		}
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

// 发送原始消息字节
func (s *Worker) sendRawMessage(client *Client, msg []byte) {
	client.Send <- msg
}

// 发送群聊消息
func (s *Worker) SendGroupMessage(fromUUID string, groupUUID string, msg []byte) {
	// 获取该群聊下的所有群成员的UUID
	groupService := &GroupService{}
	groupMembers, err := groupService.GetGroupMember(groupUUID)
	if err != nil {
		logger.Error("获取群成员失败", zap.Error(err), zap.String("groupId", groupUUID))
		return
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, clientID := range groupMembers {
		if clientID == fromUUID {
			continue
		}
		client, ok := s.Clients.Load(clientID)
		if ok {
			s.SendMessageToClient(client.(*Client), msg)
		}
	}
}
