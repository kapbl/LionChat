# 消息分片功能实现文档

## 概述

本文档描述了ChatLion聊天系统中实现的消息分片功能，该功能用于优化大消息的传输，提高系统性能和用户体验。

## 功能特性

### 核心特性
- **自动分片检测**: 自动检测超过64KB的消息并进行分片处理
- **乱序重组**: 支持分片乱序接收和正确重组
- **完整性校验**: 使用MD5校验和确保消息完整性
- **超时清理**: 自动清理超时的未完成分片
- **线程安全**: 支持高并发环境下的安全操作
- **透明处理**: 对上层应用透明，无需修改现有业务逻辑

### 技术规格
- **最大分片大小**: 64KB
- **分片超时时间**: 30秒
- **最大分片数量**: 1000个
- **校验算法**: MD5
- **支持消息类型**: 文本、文件、图片、音频、视频等

## 架构设计

### 核心组件

#### 1. FragmentManager（分片管理器）
```go
type FragmentManager struct {
    pendingMessages map[string]*FragmentInfo
    mutex           sync.RWMutex
    cleanupTicker   *time.Ticker
    done            chan struct{}
}
```

**主要职责**:
- 消息分片处理
- 分片重组管理
- 超时清理
- 并发安全控制

#### 2. FragmentInfo（分片信息）
```go
type FragmentInfo struct {
    MessageID      string
    TotalFragments int32
    Fragments      map[int32]*protocol.Message
    Timestamp      time.Time
    Checksum       string
    mutex          sync.RWMutex
}
```

**主要职责**:
- 存储分片元数据
- 管理分片集合
- 跟踪重组进度

#### 3. 扩展的Message协议
```protobuf
message Message {
    // 原有字段...
    
    // 消息分片相关字段
    string messageId = 12;      // 消息唯一标识符
    bool isFragmented = 13;     // 是否为分片消息
    int32 fragmentIndex = 14;   // 分片索引（从0开始）
    int32 totalFragments = 15;  // 总分片数
    int64 timestamp = 16;       // 消息时间戳
    string checksum = 17;       // 消息校验和
}
```

### 处理流程

#### 发送端流程
```
1. 接收原始消息
2. 检查消息大小
3. 如果 > 64KB:
   a. 生成唯一消息ID
   b. 计算消息校验和
   c. 将消息分割为多个片段
   d. 为每个片段添加元数据
   e. 依次发送所有分片
4. 如果 <= 64KB:
   a. 直接发送原始消息
```

#### 接收端流程
```
1. 接收消息片段
2. 检查是否为分片消息
3. 如果是分片:
   a. 验证分片元数据
   b. 存储分片到缓存
   c. 检查是否收集齐所有分片
   d. 如果完整，重组并返回原始消息
   e. 如果不完整，等待更多分片
4. 如果不是分片:
   a. 直接处理消息
```

## 实现细节

### 1. 消息分片算法

```go
func (fm *FragmentManager) FragmentMessage(msg *protocol.Message) ([]*protocol.Message, error) {
    // 1. 生成消息ID和校验和
    messageID := uuid.NewV4().String()
    originalBytes, _ := proto.Marshal(msg)
    checksum := fmt.Sprintf("%x", md5.Sum(originalBytes))
    
    // 2. 计算分片数量
    totalFragments := (len(originalBytes) + MaxFragmentSize - 1) / MaxFragmentSize
    
    // 3. 创建分片
    fragments := make([]*protocol.Message, 0, totalFragments)
    for i := 0; i < totalFragments; i++ {
        start := i * MaxFragmentSize
        end := start + MaxFragmentSize
        if end > len(originalBytes) {
            end = len(originalBytes)
        }
        
        fragment := &protocol.Message{
            // 复制原始消息字段
            // ...
            MessageId:      messageID,
            IsFragmented:   true,
            FragmentIndex:  int32(i),
            TotalFragments: int32(totalFragments),
            Checksum:       checksum,
            File:           originalBytes[start:end], // 分片数据
        }
        
        fragments = append(fragments, fragment)
    }
    
    return fragments, nil
}
```

### 2. 消息重组算法

```go
func (fm *FragmentManager) ProcessFragment(fragment *protocol.Message) (*protocol.Message, bool, error) {
    // 1. 验证分片
    if !fragment.IsFragmented {
        return fragment, true, nil // 非分片消息直接返回
    }
    
    // 2. 获取或创建分片信息
    fragInfo := fm.getOrCreateFragmentInfo(fragment)
    
    // 3. 存储分片
    fragInfo.Fragments[fragment.FragmentIndex] = fragment
    
    // 4. 检查是否完整
    if len(fragInfo.Fragments) == int(fragInfo.TotalFragments) {
        // 5. 重组消息
        return fm.reassembleMessage(fragInfo)
    }
    
    return nil, false, nil // 需要更多分片
}
```

### 3. 超时清理机制

```go
func (fm *FragmentManager) cleanupExpiredFragments() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            fm.mutex.Lock()
            now := time.Now()
            for messageID, fragInfo := range fm.pendingMessages {
                if now.Sub(fragInfo.Timestamp) > FragmentTimeout {
                    delete(fm.pendingMessages, messageID)
                }
            }
            fm.mutex.Unlock()
        case <-fm.done:
            return
        }
    }
}
```

## 集成方式

### 服务器端集成

1. **修改Server结构体**:
```go
type Server struct {
    // 原有字段...
    FragmentManager *FragmentManager
}
```

2. **修改消息处理逻辑**:
```go
// 在消息广播处理中添加分片处理
if msg.IsFragmented {
    completeMessage, isComplete, err := s.FragmentManager.ProcessFragment(&msg)
    if err != nil {
        // 处理错误
        continue
    }
    if !isComplete {
        // 等待更多分片
        continue
    }
    msg = *completeMessage
}
```

3. **修改消息发送逻辑**:
```go
func (s *Server) SendMessageToClient(client *Client, msg []byte) {
    var message protocol.Message
    proto.Unmarshal(msg, &message)
    
    if s.FragmentManager.ShouldFragment(&message) {
        fragments, _ := s.FragmentManager.FragmentMessage(&message)
        for _, fragment := range fragments {
            fragmentBytes, _ := proto.Marshal(fragment)
            s.sendRawMessage(client, fragmentBytes)
        }
    } else {
        s.sendRawMessage(client, msg)
    }
}
```

### 客户端集成

1. **修改Client结构体**:
```go
type Client struct {
    // 原有字段...
    FragmentManager *FragmentManager
}
```

2. **修改消息接收逻辑**:
```go
// 在客户端读取消息时添加分片处理
if pbMessage.IsFragmented {
    completeMessage, isComplete, err := c.FragmentManager.ProcessFragment(&pbMessage)
    if err != nil {
        // 处理错误
        continue
    }
    if !isComplete {
        // 等待更多分片
        continue
    }
    pbMessage = *completeMessage
}
```

## 使用示例

### 基本使用

```go
// 创建分片管理器
fm := service.NewFragmentManager()
defer fm.Stop()

// 检查是否需要分片
if fm.ShouldFragment(message) {
    // 分片消息
    fragments, err := fm.FragmentMessage(message)
    if err != nil {
        // 处理错误
        return
    }
    
    // 发送所有分片
    for _, fragment := range fragments {
        sendFragment(fragment)
    }
} else {
    // 直接发送
    sendMessage(message)
}
```

### 接收处理

```go
// 处理接收到的消息
msg, isComplete, err := fm.ProcessFragment(receivedFragment)
if err != nil {
    // 处理错误
    return
}

if isComplete {
    // 消息完整，可以处理
    processCompleteMessage(msg)
} else {
    // 等待更多分片
    // 无需额外操作，分片管理器会自动处理
}
```

## 性能优化

### 1. 内存优化
- 使用对象池减少内存分配
- 及时清理过期分片
- 限制最大分片数量

### 2. 网络优化
- 分片大小优化（64KB平衡内存和网络效率）
- 支持并行发送分片
- 智能重传机制

### 3. 并发优化
- 读写锁减少锁竞争
- 分片级别的细粒度锁
- 无锁数据结构优化

## 监控和调试

### 关键指标
- 分片消息数量
- 平均重组时间
- 超时分片数量
- 内存使用情况

### 日志记录
```go
// 分片创建日志
logger.Info("开始分片消息",
    zap.String("messageId", messageID),
    zap.Int("totalSize", totalSize),
    zap.Int("totalFragments", totalFragments))

// 重组完成日志
logger.Info("消息重组完成", 
    zap.String("messageId", messageID))

// 超时清理日志
logger.Warn("清理过期分片",
    zap.String("messageId", messageID),
    zap.Duration("age", age))
```

## 测试

### 单元测试
- 分片功能测试
- 重组功能测试
- 乱序处理测试
- 超时清理测试
- 错误处理测试

### 集成测试
- 端到端消息传输测试
- 高并发场景测试
- 网络异常场景测试

### 性能测试
- 大消息传输性能
- 内存使用测试
- 并发处理能力测试

## 运行演示

```bash
# 运行分片功能演示
go run examples/fragment_demo.go

# 运行单元测试
go test ./internal/service -v -run TestFragmentManager

# 运行性能测试
go test ./internal/service -bench=BenchmarkFragment
```

## 配置参数

```go
const (
    MaxFragmentSize = 64 * 1024    // 最大分片大小 (64KB)
    FragmentTimeout = 30 * time.Second // 分片超时时间 (30秒)
    MaxFragments = 1000             // 最大分片数量
)
```

## 注意事项

1. **内存管理**: 大消息会占用较多内存，需要合理控制并发数量
2. **网络稳定性**: 分片传输依赖网络稳定性，建议配合重传机制
3. **顺序保证**: 虽然支持乱序接收，但建议按顺序发送以提高效率
4. **错误处理**: 需要妥善处理分片丢失、超时等异常情况
5. **版本兼容**: 新旧客户端需要保持协议兼容性

## 未来改进

1. **压缩支持**: 在分片前对消息进行压缩
2. **加密支持**: 对分片进行端到端加密
3. **自适应分片**: 根据网络状况动态调整分片大小
4. **重传机制**: 实现分片级别的重传
5. **流式传输**: 支持流式大文件传输