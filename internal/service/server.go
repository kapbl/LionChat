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
	Clients   sync.Map     // 存储所有客户端连接
	mutex     *sync.Mutex  // 保护clients的并发访问
	Broadcast chan []byte  // 广播通道
	Register  chan *Client // 注册消息通道
	Ungister  chan *Client // 注销消息通道
}

var ServerInstance = &Server{
	Clients:   sync.Map{},
	mutex:     &sync.Mutex{},
	Broadcast: make(chan []byte, 1000), // 1000缓冲大小
	Register:  make(chan *Client, 100),
	Ungister:  make(chan *Client, 100),
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
			s.mutex.Lock()
			logger.Info("注销连接", zap.String("uuid", conn.UUID))
			s.Clients.Delete(conn.UUID)
			close(conn.Send)
			// 发送用户下线事件到Kafka
			if dao.KafkaProducerInstance != nil {
				metadata := map[string]interface{}{
					"disconnect_time": conn.ConnTime,
					"client_ip":       conn.RemoteAddr,
				}
				if err := dao.KafkaProducerInstance.SendUserEvent("user_offline", conn.UUID, metadata); err != nil {
					logger.Error("发送用户下线事件到Kafka失败", zap.Error(err))
				}
			}
			s.mutex.Unlock()
		case message := <-s.Broadcast:
			msg := protocol.Message{}
			if err := proto.Unmarshal(message, &msg); err != nil {
				logger.Error("解析消息失败", zap.Error(err))
				continue
			}
			// 表示有意向的消息
			if msg.To != "" {
				logger.Info("处理定向消息",
					zap.String("from", msg.From),
					zap.String("to", msg.To),
					zap.Int32("contentType", msg.ContentType),
					zap.Int32("messageType", msg.MessageType))
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
							msgByte, err := proto.Marshal(&msg)
							if err != nil {
								logger.Error("消息序列化失败", zap.Error(err))
								return
							}
							logger.Info("发送单聊文件消息", zap.String("to", msg.To))
							s.SendMessageToClient(client.(*Client), msgByte)
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
							msgByte, err := proto.Marshal(&msg)
							if err != nil {
								logger.Error("消息序列化失败", zap.Error(err))
								return
							}
							logger.Info("发送单聊图片消息", zap.String("to", msg.To))
							s.SendMessageToClient(client.(*Client), msgByte)
						}
					} else {
						// 群聊消息
						msgByte, err := proto.Marshal(&msg)
						if err != nil {
							logger.Error("消息序列化失败", zap.Error(err))
							return
						}
						logger.Info("发送群聊图片消息", zap.String("groupId", msg.To))
						s.SendGroupMessage(msg.From, msg.To, msgByte)
					}
				case 4: // 语音消息
					if msg.MessageType == 1 {
						client, ok := s.Clients.Load(msg.To)
						if ok {
							msgByte, err := proto.Marshal(&msg)
							if err != nil {
								logger.Error("消息序列化失败", zap.Error(err))
								return
							}
							logger.Info("发送单聊语音消息", zap.String("to", msg.To))
							s.SendMessageToClient(client.(*Client), msgByte)
						}
					} else {
						// 群聊消息
						msgByte, err := proto.Marshal(&msg)
						if err != nil {
							logger.Error("消息序列化失败", zap.Error(err))
							return
						}
						logger.Info("发送群聊语音消息", zap.String("groupId", msg.To))
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

// 发送单个消息
func (s *Server) SendMessageToClient(client *Client, msg []byte) {
	// 先检查客户端是否存在，避免在发送消息时才发现客户端已断开
	select {
	case client.Send <- msg:
		logger.Debug("消息发送成功", zap.String("to", client.UUID))
	case <-client.done:
		logger.Info("客户端已关闭", zap.String("uuid", client.UUID))
		s.handleClientDisconnect(client)
	default:
		logger.Warn("消息发送失败，客户端可能已断开", zap.String("uuid", client.UUID))
		s.handleClientDisconnect(client)
	}
}

func (s *Server) handleClientDisconnect(client *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	close(client.Send)
	s.Clients.Delete(client.UUID)
}
