package service

import (
	"cchat/internal/dao"
	"cchat/pkg/logger"
	"cchat/pkg/protocol"
	"sync"

	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"
)

type Server struct {
	Clients         sync.Map         // 存储所有客户端连接
	mutex           *sync.Mutex      // 保护clients的并发访问
	Broadcast       chan []byte      // 广播通道
	Register        chan *Client     // 注册消息通道
	Ungister        chan *Client     // 注销消息通道
	FragmentManager *FragmentManager // 消息分片管理器
}

var ServerInstance = &Server{
	Clients:         sync.Map{},
	mutex:           &sync.Mutex{},
	Broadcast:       make(chan []byte, 1000), // 1000缓冲大小
	Register:        make(chan *Client, 100),
	Ungister:        make(chan *Client, 100),
	FragmentManager: NewFragmentManager(),
}

// 服务启动入口
func (s *Server) Start() {
	for {
		select {
		case conn := <-s.Register:
			s.mutex.Lock()
			logger.Info("注册连接", zap.String("uuid", conn.UUID))
			s.Clients.Store(conn.UUID, conn)
			// 发送用户上线事件到Kafka
			if dao.KafkaProducerInstance != nil {
				metadata := map[string]interface{}{
					"connection_time": conn.ConnTime,
					"client_ip":       conn.RemoteAddr,
				}
				if err := dao.KafkaProducerInstance.SendUserEvent("user_online", conn.UUID, metadata); err != nil {
					logger.Error("发送用户上线事件到Kafka失败", zap.Error(err))
				}
			}
			s.mutex.Unlock()
		case conn := <-s.Ungister:
			s.handleClientDisconnect(conn) // 统一使用清理函数
		case message := <-s.Broadcast:
			msg := protocol.Message{}
			if err := proto.Unmarshal(message, &msg); err != nil {
				logger.Error("解析消息失败", zap.Error(err))
				continue
			}
			// 表示有意向的消息
			if msg.To != "" {
				s.mutex.Lock()
				switch msg.ContentType {
				case 1: //Text消息
					if msg.MessageType == 1 {
						// 查找这个客户端
						client, ok := s.Clients.Load(msg.To)
						if ok {
							msgByte, err := proto.Marshal(&msg)
							if err != nil {
								logger.Error("消息序列化失败", zap.Error(err))
								s.mutex.Unlock()
								continue
							}
							logger.Info("发送单聊文本消息", zap.String("to", msg.To))
							// 发送消息 Kafka路径
							if dao.KafkaProducerInstance != nil {
								if err := dao.KafkaProducerInstance.SendChatMessage(msg.To, msgByte); err != nil {
									logger.Error("发送消息到Kafka失败", zap.Error(err))
								}
							} else { // 发送消息 websockt 路径
								s.SendMessageToClient(client.(*Client), msgByte)
							}
							s.mutex.Unlock()
						} else {
							s.mutex.Unlock()
						}
					} else {
						// 群聊消息
						msgByte, err := proto.Marshal(&msg)
						if err != nil {
							logger.Error("消息序列化失败", zap.Error(err))
							s.mutex.Unlock()
							continue
						}
						logger.Info("发送群聊文本消息", zap.String("groupId", msg.To))

						// 发送群聊消息到Kafka
						if dao.KafkaProducerInstance != nil {
							if err := dao.KafkaProducerInstance.SendGroupMessage(msg.To, msg.From, msgByte); err != nil {
								logger.Error("发送群聊消息到Kafka失败", zap.Error(err))
							}
						} else {
							s.SendGroupMessage(msg.From, msg.To, msgByte)
						}

						s.mutex.Unlock()
					}
				case 2: // 文件消息
					if msg.MessageType == 1 {
						client, ok := s.Clients.Load(msg.To)
						if ok {
							// 检查消息的大小
							if s.FragmentManager.ShouldFragment(&msg) {
								// 分片消息
								fragments, err := s.FragmentManager.FragmentMessage(&msg)
								if err != nil {
									logger.Error("消息分片失败", zap.Error(err))
									s.mutex.Unlock()
									return
								}
								// 发送分片消息
								for _, fragment := range fragments {
									fragmentBytes, err := proto.Marshal(fragment)
									if err != nil {
										logger.Error("分片序列化失败", zap.Error(err))
										s.mutex.Unlock()
										return
									}
									client.(*Client).Send <- fragmentBytes
								}
							}
							s.mutex.Unlock()
						}
					} else {
						// 群聊消息
						msgByte, err := proto.Marshal(&msg)
						if err != nil {
							logger.Error("消息序列化失败", zap.Error(err))
							return
						}
						logger.Info("发送群聊文件消息", zap.String("groupId", msg.To))
						s.SendGroupMessage(msg.From, msg.To, msgByte)
					}
				case 3: // 图片消息
					if msg.MessageType == 1 {
						client, ok := s.Clients.Load(msg.To)
						if ok {
							// 检查消息的大小
							if s.FragmentManager.ShouldFragment(&msg) {
								// 分片消息
								fragments, err := s.FragmentManager.FragmentMessage(&msg)
								if err != nil {
									logger.Error("消息分片失败", zap.Error(err))
									s.mutex.Unlock()
									return
								}
								// 发送分片消息
								for _, fragment := range fragments {
									fragmentBytes, err := proto.Marshal(fragment)
									if err != nil {
										logger.Error("分片序列化失败", zap.Error(err))
										s.mutex.Unlock()
										return
									}
									client.(*Client).Send <- fragmentBytes
								}
							}
							s.mutex.Unlock()
						}
					} else {
						// 群聊消息
						msgByte, err := proto.Marshal(&msg)
						if err != nil {
							logger.Error("消息序列化失败", zap.Error(err))
							return
						}
						logger.Info("发送群聊文件消息", zap.String("groupId", msg.To))
						s.SendGroupMessage(msg.From, msg.To, msgByte)
					}
				case 5: // 视频消息
					if msg.MessageType == 1 {
						client, ok := s.Clients.Load(msg.To)
						if ok {
							// 检查消息的大小
							if s.FragmentManager.ShouldFragment(&msg) {
								// 分片消息
								fragments, err := s.FragmentManager.FragmentMessage(&msg)
								if err != nil {
									logger.Error("消息分片失败", zap.Error(err))
									s.mutex.Unlock()
									return
								}
								// 发送分片消息
								for _, fragment := range fragments {
									fragmentBytes, err := proto.Marshal(fragment)
									if err != nil {
										logger.Error("分片序列化失败", zap.Error(err))
										s.mutex.Unlock()
										return
									}
									client.(*Client).Send <- fragmentBytes
								}
							}
							s.mutex.Unlock()
						}
					} else {
						// 群聊消息
						msgByte, err := proto.Marshal(&msg)
						if err != nil {
							logger.Error("消息序列化失败", zap.Error(err))
							return
						}
						logger.Info("发送群聊文件消息", zap.String("groupId", msg.To))
						s.SendGroupMessage(msg.From, msg.To, msgByte)
					}
				case 4: // 语音消息
					if msg.MessageType == 1 {
						client, ok := s.Clients.Load(msg.To)
						if ok {
							// 检查消息的大小
							if s.FragmentManager.ShouldFragment(&msg) {
								// 分片消息
								fragments, err := s.FragmentManager.FragmentMessage(&msg)
								if err != nil {
									logger.Error("消息分片失败", zap.Error(err))
									s.mutex.Unlock()
									return
								}
								// 发送分片消息
								for _, fragment := range fragments {
									fragmentBytes, err := proto.Marshal(fragment)
									if err != nil {
										logger.Error("分片序列化失败", zap.Error(err))
										s.mutex.Unlock()
										return
									}
									client.(*Client).Send <- fragmentBytes
								}
							} else {
								// 发送单聊语音消息
								logger.Info("发送单聊语音消息", zap.String("to", msg.To))
								msgByte, err := proto.Marshal(&msg)
								if err != nil {
									logger.Error("消息序列化失败", zap.Error(err))
									return
								}
								s.SendMessageToClient(client.(*Client), msgByte)
							}
							s.mutex.Unlock()
						}
					} else {
						// 群聊消息
						msgByte, err := proto.Marshal(&msg)
						if err != nil {
							logger.Error("消息序列化失败", zap.Error(err))
							return
						}
						logger.Info("发送群聊文件消息", zap.String("groupId", msg.To))
						s.SendGroupMessage(msg.From, msg.To, msgByte)
					}
				case 8: // 加好友消息
					if msg.MessageType == 1 {
						client, ok := s.Clients.Load(msg.To)
						if ok {
							msgByte, err := proto.Marshal(&msg)
							if err != nil {
								logger.Error("消息序列化失败", zap.Error(err))
								return
							}
							logger.Info("发送好友请求消息", zap.String("to", msg.To))
							s.SendMessageToClient(client.(*Client), msgByte)
						}
					}
				}
			} else {
				logger.Info("发送系统公告消息")
				s.mutex.Lock()
				s.Clients.Range(func(key, value interface{}) bool {
					client := value.(*Client)
					s.SendMessageToClient(client, message)
					return true
				})
				s.mutex.Unlock()
			}
		}
	}
}

// 使用 uuid 获取客户端
func (s *Server) GetClient(uuid string) *Client {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	client, ok := s.Clients.Load(uuid)
	if ok {
		return client.(*Client)
	}
	return nil
}

// 发送群聊消息
func (s *Server) SendGroupMessage(fromUUID string, groupUUID string, msg []byte) {
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

// 发送单个消息（支持分片）
func (s *Server) SendMessageToClient(client *Client, msg []byte) {
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
		for i, fragment := range fragments {
			fragmentBytes, err := proto.Marshal(fragment)
			if err != nil {
				logger.Error("分片序列化失败", zap.Error(err))
				return
			}

			if !s.sendRawMessage(client, fragmentBytes) {
				logger.Error("发送分片失败",
					zap.String("to", client.UUID),
					zap.Int("fragmentIndex", i))
				return
			}
		}

		logger.Info("分片消息发送完成",
			zap.String("to", client.UUID),
			zap.Int("totalFragments", len(fragments)))
	} else {
		// 直接发送非分片消息
		s.sendRawMessage(client, msg)
	}
}

// 发送原始消息字节
func (s *Server) sendRawMessage(client *Client, msg []byte) bool {
	select {
	case client.Send <- msg:
		return true
	case <-client.done:
		// 不直接调用handleClientDisconnect，让Ungister流程处理
		logger.Info("检测到客户端已关闭", zap.String("uuid", client.UUID))
		return false
	default:
		logger.Warn("消息发送失败，客户端可能已断开", zap.String("uuid", client.UUID))
		return false
	}
}

func (s *Server) handleClientDisconnect(client *Client) {
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

	// 清理分片管理器
	if client.FragmentManager != nil {
		client.FragmentManager.Stop()
	}

	// 发送Kafka下线事件
	if dao.KafkaProducerInstance != nil {
		// ... Kafka事件发送逻辑
	}

	logger.Info("客户端连接已清理", zap.String("uuid", client.UUID))
}
