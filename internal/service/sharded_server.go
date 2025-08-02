package service

import (
	"cchat/internal/dao"
	"cchat/pkg/logger"
	"cchat/pkg/protocol"
	"hash/fnv"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"
)

// MessageShard 消息分片，减少锁竞争
type MessageShard struct {
	clients   sync.Map
	mutex     sync.RWMutex
	broadcast chan []byte
	register  chan *Client
	unregister chan *Client
	shardID   int
}

// ShardedServer 分片服务器，提高并发性能
type ShardedServer struct {
	shards   []*MessageShard
	shardNum int
	wg       sync.WaitGroup
}

// NewShardedServer 创建分片服务器
func NewShardedServer(shardNum int) *ShardedServer {
	server := &ShardedServer{
		shards:   make([]*MessageShard, shardNum),
		shardNum: shardNum,
	}

	// 初始化每个分片
	for i := 0; i < shardNum; i++ {
		server.shards[i] = &MessageShard{
			clients:    sync.Map{},
			mutex:      sync.RWMutex{},
			broadcast:  make(chan []byte, 2000), // 增加缓冲区大小
			register:   make(chan *Client, 200),
			unregister: make(chan *Client, 200),
			shardID:    i,
		}
	}

	return server
}

// hashUserID 根据用户ID计算分片索引
func (s *ShardedServer) hashUserID(userID string) int {
	h := fnv.New32a()
	h.Write([]byte(userID))
	return int(h.Sum32()) % s.shardNum
}

// GetShard 获取用户对应的分片
func (s *ShardedServer) GetShard(userID string) *MessageShard {
	shardIndex := s.hashUserID(userID)
	return s.shards[shardIndex]
}

// Start 启动所有分片
func (s *ShardedServer) Start() {
	for i, shard := range s.shards {
		s.wg.Add(1)
		go func(shardID int, shard *MessageShard) {
			defer s.wg.Done()
			shard.start(shardID)
		}(i, shard)
	}
	logger.Info("分片服务器启动完成", zap.Int("shardNum", s.shardNum))
}

// Stop 停止所有分片
func (s *ShardedServer) Stop() {
	for _, shard := range s.shards {
		close(shard.broadcast)
		close(shard.register)
		close(shard.unregister)
	}
	s.wg.Wait()
	logger.Info("分片服务器已停止")
}

// RegisterClient 注册客户端到对应分片
func (s *ShardedServer) RegisterClient(client *Client) {
	shard := s.GetShard(client.UUID)
	select {
	case shard.register <- client:
		logger.Debug("客户端注册到分片", 
			zap.String("uuid", client.UUID),
			zap.Int("shardID", shard.shardID))
	case <-time.After(5 * time.Second):
		logger.Error("客户端注册超时", zap.String("uuid", client.UUID))
	}
}

// UnregisterClient 从对应分片注销客户端
func (s *ShardedServer) UnregisterClient(client *Client) {
	shard := s.GetShard(client.UUID)
	select {
	case shard.unregister <- client:
		logger.Debug("客户端从分片注销", 
			zap.String("uuid", client.UUID),
			zap.Int("shardID", shard.shardID))
	case <-time.After(5 * time.Second):
		logger.Error("客户端注销超时", zap.String("uuid", client.UUID))
	}
}

// BroadcastMessage 广播消息到对应分片
func (s *ShardedServer) BroadcastMessage(userID string, message []byte) {
	shard := s.GetShard(userID)
	select {
	case shard.broadcast <- message:
		logger.Debug("消息发送到分片", 
			zap.String("userID", userID),
			zap.Int("shardID", shard.shardID))
	case <-time.After(3 * time.Second):
		logger.Error("消息广播超时", 
			zap.String("userID", userID),
			zap.Int("shardID", shard.shardID))
	default:
		logger.Error("分片广播通道已满", 
			zap.String("userID", userID),
			zap.Int("shardID", shard.shardID))
	}
}

// GetClient 从对应分片获取客户端
func (s *ShardedServer) GetClient(userID string) (*Client, bool) {
	shard := s.GetShard(userID)
	shard.mutex.RLock()
	defer shard.mutex.RUnlock()
	
	if client, ok := shard.clients.Load(userID); ok {
		return client.(*Client), true
	}
	return nil, false
}

// start 启动单个分片的消息处理循环
func (shard *MessageShard) start(shardID int) {
	logger.Info("分片启动", zap.Int("shardID", shardID))
	
	for {
		select {
		case client := <-shard.register:
			if client == nil {
				return
			}
			shard.handleClientRegister(client)
			
		case client := <-shard.unregister:
			if client == nil {
				return
			}
			shard.handleClientUnregister(client)
			
		case message := <-shard.broadcast:
			if message == nil {
				return
			}
			shard.handleMessage(message)
		}
	}
}

// handleClientRegister 处理客户端注册
func (shard *MessageShard) handleClientRegister(client *Client) {
	shard.mutex.Lock()
	defer shard.mutex.Unlock()
	
	logger.Info("分片注册客户端", 
		zap.String("uuid", client.UUID),
		zap.Int("shardID", shard.shardID))
	
	shard.clients.Store(client.UUID, client)
	
	// 发送用户上线事件到Kafka
	if dao.KafkaProducerInstance != nil {
		metadata := map[string]interface{}{
			"connection_time": client.ConnTime,
			"client_ip":       client.RemoteAddr,
			"shard_id":        shard.shardID,
		}
		if err := dao.KafkaProducerInstance.SendUserEvent("user_online", client.UUID, metadata); err != nil {
			logger.Error("发送用户上线事件到Kafka失败", zap.Error(err))
		}
	}
}

// handleClientUnregister 处理客户端注销
func (shard *MessageShard) handleClientUnregister(client *Client) {
	shard.mutex.Lock()
	defer shard.mutex.Unlock()
	
	logger.Info("分片注销客户端", 
		zap.String("uuid", client.UUID),
		zap.Int("shardID", shard.shardID))
	
	shard.clients.Delete(client.UUID)
	close(client.Send)
	
	// 发送用户下线事件到Kafka
	if dao.KafkaProducerInstance != nil {
		metadata := map[string]interface{}{
			"disconnect_time": time.Now().Unix(),
			"client_ip":       client.RemoteAddr,
			"shard_id":        shard.shardID,
		}
		if err := dao.KafkaProducerInstance.SendUserEvent("user_offline", client.UUID, metadata); err != nil {
			logger.Error("发送用户下线事件到Kafka失败", zap.Error(err))
		}
	}
}

// handleMessage 处理消息广播
func (shard *MessageShard) handleMessage(message []byte) {
	msg := protocol.Message{}
	if err := proto.Unmarshal(message, &msg); err != nil {
		logger.Error("解析消息失败", zap.Error(err))
		return
	}
	
	// 处理定向消息
	if msg.To != "" {
		shard.handleDirectMessage(&msg, message)
	} else {
		// 处理广播消息（如果需要）
		shard.handleBroadcastMessage(&msg, message)
	}
}

// handleDirectMessage 处理定向消息
func (shard *MessageShard) handleDirectMessage(msg *protocol.Message, rawMessage []byte) {
	shard.mutex.RLock()
	client, ok := shard.clients.Load(msg.To)
	shard.mutex.RUnlock()
	
	if !ok {
		logger.Debug("目标客户端不在此分片", 
			zap.String("to", msg.To),
			zap.Int("shardID", shard.shardID))
		return
	}
	
	logger.Info("分片处理定向消息",
		zap.String("from", msg.From),
		zap.String("to", msg.To),
		zap.Int32("contentType", msg.ContentType),
		zap.Int32("messageType", msg.MessageType),
		zap.Int("shardID", shard.shardID))
	
	// 根据消息类型处理
	switch msg.ContentType {
	case 1: // 文本消息
		if msg.MessageType == 1 {
			// 单聊消息
			msgByte, err := proto.Marshal(msg)
			if err != nil {
				logger.Error("消息序列化失败", zap.Error(err))
				return
			}
			
			// 优先使用Kafka
			if dao.KafkaProducerInstance != nil {
				if err := dao.KafkaProducerInstance.SendChatMessage(msg.To, msgByte); err != nil {
					logger.Error("发送消息到Kafka失败", zap.Error(err))
					// Kafka失败时降级到WebSocket
					shard.sendMessageToClient(client.(*Client), msgByte)
				}
			} else {
				// 直接通过WebSocket发送
				shard.sendMessageToClient(client.(*Client), msgByte)
			}
		} else {
			// 群聊消息 - 这里需要跨分片处理
			shard.handleGroupMessage(msg)
		}
	case 2, 3, 4: // 文件、图片、语音消息
		msgByte, err := proto.Marshal(msg)
		if err != nil {
			logger.Error("消息序列化失败", zap.Error(err))
			return
		}
		shard.sendMessageToClient(client.(*Client), msgByte)
	}
}

// handleBroadcastMessage 处理广播消息
func (shard *MessageShard) handleBroadcastMessage(msg *protocol.Message, rawMessage []byte) {
	// 向分片内所有客户端广播
	shard.mutex.RLock()
	defer shard.mutex.RUnlock()
	
	msgByte, err := proto.Marshal(msg)
	if err != nil {
		logger.Error("消息序列化失败", zap.Error(err))
		return
	}
	
	shard.clients.Range(func(key, value interface{}) bool {
		client := value.(*Client)
		if client.UUID != msg.From { // 不发送给自己
			shard.sendMessageToClient(client, msgByte)
		}
		return true
	})
}

// handleGroupMessage 处理群聊消息（需要跨分片协调）
func (shard *MessageShard) handleGroupMessage(msg *protocol.Message) {
	// 群聊消息需要通过Kafka处理，因为成员可能分布在不同分片
	msgByte, err := proto.Marshal(msg)
	if err != nil {
		logger.Error("群聊消息序列化失败", zap.Error(err))
		return
	}
	
	if dao.KafkaProducerInstance != nil {
		if err := dao.KafkaProducerInstance.SendGroupMessage(msg.To, msg.From, msgByte); err != nil {
			logger.Error("发送群聊消息到Kafka失败", zap.Error(err))
		}
	} else {
		logger.Warn("Kafka未配置，无法处理群聊消息")
	}
}

// sendMessageToClient 发送消息给客户端（优化版本）
func (shard *MessageShard) sendMessageToClient(client *Client, msg []byte) {
	select {
	case client.Send <- msg:
		logger.Debug("消息发送成功", 
			zap.String("to", client.UUID),
			zap.Int("shardID", shard.shardID))
	case <-client.done:
		logger.Info("客户端已关闭", zap.String("uuid", client.UUID))
		shard.handleClientDisconnect(client)
	case <-time.After(2 * time.Second):
		logger.Warn("消息发送超时", 
			zap.String("uuid", client.UUID),
			zap.Int("shardID", shard.shardID))
		shard.handleClientDisconnect(client)
	default:
		logger.Warn("客户端发送队列已满", 
			zap.String("uuid", client.UUID),
			zap.Int("shardID", shard.shardID))
		// 可以考虑丢弃消息或者进行其他处理
	}
}

// handleClientDisconnect 处理客户端断开连接
func (shard *MessageShard) handleClientDisconnect(client *Client) {
	shard.mutex.Lock()
	defer shard.mutex.Unlock()
	
	if _, exists := shard.clients.Load(client.UUID); exists {
		shard.clients.Delete(client.UUID)
		close(client.Send)
		logger.Info("清理断开的客户端", 
			zap.String("uuid", client.UUID),
			zap.Int("shardID", shard.shardID))
	}
}

// GetShardStats 获取分片统计信息
func (shard *MessageShard) GetShardStats() map[string]interface{} {
	shard.mutex.RLock()
	defer shard.mutex.RUnlock()
	
	clientCount := 0
	shard.clients.Range(func(key, value interface{}) bool {
		clientCount++
		return true
	})
	
	return map[string]interface{}{
		"shard_id":      shard.shardID,
		"client_count":  clientCount,
		"broadcast_cap":  cap(shard.broadcast),
		"broadcast_len":  len(shard.broadcast),
		"register_cap":   cap(shard.register),
		"register_len":   len(shard.register),
		"unregister_cap": cap(shard.unregister),
		"unregister_len": len(shard.unregister),
	}
}

// GetServerStats 获取整个服务器的统计信息
func (s *ShardedServer) GetServerStats() map[string]interface{} {
	totalClients := 0
	shardStats := make([]map[string]interface{}, len(s.shards))
	
	for i, shard := range s.shards {
		stats := shard.GetShardStats()
		shardStats[i] = stats
		totalClients += stats["client_count"].(int)
	}
	
	return map[string]interface{}{
		"total_clients": totalClients,
		"shard_count":   s.shardNum,
		"shard_stats":   shardStats,
	}
}