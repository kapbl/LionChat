package service

import (
	"cchat/pkg/logger"
	"cchat/pkg/protocol"
	"crypto/md5"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

const (
	// 最大分片大小 (64KB)
	MaxFragmentSize = 64 * 1024
	// 分片超时时间 (30秒)
	FragmentTimeout = 30 * time.Second
	// 最大分片数量
	MaxFragments = 1000
)

// FragmentInfo 分片信息
type FragmentInfo struct {
	MessageID      string
	TotalFragments int32
	Fragments      map[int32]*protocol.Message
	Timestamp      time.Time
	Checksum       string
	mutex          sync.RWMutex
}

// FragmentManager 消息分片管理器
type FragmentManager struct {
	pendingMessages map[string]*FragmentInfo
	mutex           sync.RWMutex
	cleanupTicker   *time.Ticker
	done            chan struct{}
}

// NewFragmentManager 创建新的分片管理器
func NewFragmentManager() *FragmentManager {
	fm := &FragmentManager{
		pendingMessages: make(map[string]*FragmentInfo),
		cleanupTicker:   time.NewTicker(time.Minute),
		done:            make(chan struct{}),
	}

	// 启动清理协程
	go fm.cleanupExpiredFragments()

	return fm
}

// Stop 停止分片管理器
func (fm *FragmentManager) Stop() {
	close(fm.done)
	fm.cleanupTicker.Stop()
}

// ShouldFragment 判断消息是否需要分片
func (fm *FragmentManager) ShouldFragment(msg *protocol.Message) bool {
	// 序列化消息以获取实际大小
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return false
	}

	return len(msgBytes) > MaxFragmentSize
}

// FragmentMessage 将消息分片
func (fm *FragmentManager) FragmentMessage(msg *protocol.Message) ([]*protocol.Message, error) {
	// 生成消息ID
	messageID := uuid.NewV4().String()
	msg.MessageId = messageID
	msg.Timestamp = time.Now().Unix()

	// 序列化原始消息
	originalBytes, err := proto.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("序列化消息失败: %v", err)
	}

	// 计算校验和
	checksum := fmt.Sprintf("%x", md5.Sum(originalBytes))
	msg.Checksum = checksum

	// 计算需要的分片数量
	totalSize := len(originalBytes)
	logger.Info("原始消息大小", zap.Int("totalSize", totalSize))
	totalFragments := (totalSize + MaxFragmentSize - 1) / MaxFragmentSize

	if totalFragments > MaxFragments {
		return nil, fmt.Errorf("消息过大，分片数量超过限制: %d > %d", totalFragments, MaxFragments)
	}

	logger.Info("开始分片消息",
		zap.String("messageId", messageID),
		zap.Int("totalSize", totalSize),
		zap.Int("totalFragments", totalFragments))

	fragments := make([]*protocol.Message, 0, totalFragments)

	// 创建分片
	for i := 0; i < totalFragments; i++ {
		start := i * MaxFragmentSize
		end := start + MaxFragmentSize
		if end > totalSize {
			end = totalSize
		}

		// 创建分片消息
		fragment := &protocol.Message{
			Avatar:         msg.Avatar,
			FromUsername:   msg.FromUsername,
			From:           msg.From,
			To:             msg.To,
			Content:        msg.Content,
			ContentType:    msg.ContentType,
			Type:           msg.Type,
			MessageType:    msg.MessageType,
			FileSuffix:     msg.FileSuffix,
			MessageId:      messageID,
			IsFragmented:   true,
			FragmentIndex:  int32(i),
			TotalFragments: int32(totalFragments),
			Timestamp:      msg.Timestamp,
			Checksum:       checksum,
			File:           originalBytes[start:end], // 使用file字段存储分片数据
		}

		fragments = append(fragments, fragment)
	}

	return fragments, nil
}

// ProcessFragment 处理接收到的分片
func (fm *FragmentManager) ProcessFragment(fragment *protocol.Message) (*protocol.Message, bool, error) {
	if !fragment.IsFragmented {
		return fragment, true, nil // 非分片消息直接返回
	}

	messageID := fragment.MessageId
	if messageID == "" {
		return nil, false, fmt.Errorf("分片消息缺少messageId")
	}

	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	// 获取或创建分片信息
	fragInfo, exists := fm.pendingMessages[messageID]
	if !exists {
		fragInfo = &FragmentInfo{
			MessageID:      messageID,
			TotalFragments: fragment.TotalFragments,
			Fragments:      make(map[int32]*protocol.Message),
			Timestamp:      time.Now(),
			Checksum:       fragment.Checksum,
		}
		fm.pendingMessages[messageID] = fragInfo
	}

	fragInfo.mutex.Lock()
	defer fragInfo.mutex.Unlock()

	// 验证分片信息
	if fragment.TotalFragments != fragInfo.TotalFragments {
		return nil, false, fmt.Errorf("分片总数不匹配: %d != %d", fragment.TotalFragments, fragInfo.TotalFragments)
	}

	if fragment.Checksum != fragInfo.Checksum {
		return nil, false, fmt.Errorf("校验和不匹配")
	}

	// 存储分片
	fragInfo.Fragments[fragment.FragmentIndex] = fragment

	logger.Debug("接收到分片",
		zap.String("messageId", messageID),
		zap.Int32("fragmentIndex", fragment.FragmentIndex),
		zap.Int("receivedFragments", len(fragInfo.Fragments)),
		zap.Int32("totalFragments", fragInfo.TotalFragments))

	// 检查是否收集齐所有分片
	if int32(len(fragInfo.Fragments)) == fragInfo.TotalFragments {
		// 重组消息
		completeMessage, err := fm.reassembleMessage(fragInfo)
		if err != nil {
			return nil, false, fmt.Errorf("重组消息失败: %v", err)
		}

		// 清理分片信息
		delete(fm.pendingMessages, messageID)

		logger.Info("消息重组完成", zap.String("messageId", messageID))
		return completeMessage, true, nil
	}

	return nil, false, nil // 还需要更多分片
}

// reassembleMessage 重组消息
func (fm *FragmentManager) reassembleMessage(fragInfo *FragmentInfo) (*protocol.Message, error) {
	// 按分片索引排序
	indices := make([]int32, 0, len(fragInfo.Fragments))
	for index := range fragInfo.Fragments {
		indices = append(indices, index)
	}
	sort.Slice(indices, func(i, j int) bool {
		return indices[i] < indices[j]
	})

	// 重组数据
	var reassembledData []byte
	for _, index := range indices {
		fragment := fragInfo.Fragments[index]
		reassembledData = append(reassembledData, fragment.File...)
	}

	// 验证校验和
	actualChecksum := fmt.Sprintf("%x", md5.Sum(reassembledData))
	if actualChecksum != fragInfo.Checksum {
		return nil, fmt.Errorf("重组后校验和不匹配: %s != %s", actualChecksum, fragInfo.Checksum)
	}

	// 反序列化原始消息
	var originalMessage protocol.Message
	if err := proto.Unmarshal(reassembledData, &originalMessage); err != nil {
		return nil, fmt.Errorf("反序列化重组消息失败: %v", err)
	}

	// 重置分片相关字段
	originalMessage.IsFragmented = false
	originalMessage.FragmentIndex = 0
	originalMessage.TotalFragments = 0

	return &originalMessage, nil
}

// cleanupExpiredFragments 清理过期的分片
func (fm *FragmentManager) cleanupExpiredFragments() {
	for {
		select {
		case <-fm.cleanupTicker.C:
			fm.mutex.Lock()
			now := time.Now()
			for messageID, fragInfo := range fm.pendingMessages {
				if now.Sub(fragInfo.Timestamp) > FragmentTimeout {
					logger.Warn("清理过期分片",
						zap.String("messageId", messageID),
						zap.Duration("age", now.Sub(fragInfo.Timestamp)))
					delete(fm.pendingMessages, messageID)
				}
			}
			fm.mutex.Unlock()
		case <-fm.done:
			return
		}
	}
}

// GetPendingFragmentsCount 获取待处理分片数量
func (fm *FragmentManager) GetPendingFragmentsCount() int {
	fm.mutex.RLock()
	defer fm.mutex.RUnlock()
	return len(fm.pendingMessages)
}
