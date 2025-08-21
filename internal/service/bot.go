package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Bot struct {
	Client
	Username string
	UUID     string
	UserID   int64
	Context  []string
}

// DeepSeek API 请求结构
type DeepSeekRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeek API 响应结构
type DeepSeekResponse struct {
	Choices []Choice  `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

type Choice struct {
	Message Message `json:"message"`
}

type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// 接入DeepSeek API
func (b *Bot) GetBotResponse(prompt string) (string, error) {
	// 获取API密钥
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("DEEPSEEK_API_KEY environment variable not set")
	}

	// 构建消息历史
	messages := []Message{
		{Role: "system", Content: "You are a helpful assistant."},
	}

	// 添加历史上下文
	for i, context := range b.Context {
		if i%2 == 0 {
			// 用户消息
			messages = append(messages, Message{Role: "user", Content: context})
		} else {
			// 助手回复
			messages = append(messages, Message{Role: "assistant", Content: context})
		}
	}

	// 添加当前用户消息
	messages = append(messages, Message{Role: "user", Content: prompt})

	// 构建请求
	request := DeepSeekRequest{
		Model:    "deepseek-chat",
		Messages: messages,
		Stream:   false,
	}

	// 序列化请求
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", "https://api.deepseek.com/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response DeepSeekResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	// 检查API错误
	if response.Error != nil {
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	// 检查响应格式
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	// 获取回复内容
	botReply := response.Choices[0].Message.Content

	// 更新上下文
	b.Context = append(b.Context, prompt, botReply)

	// 限制上下文长度，保留最近的10轮对话
	if len(b.Context) > 20 {
		b.Context = b.Context[len(b.Context)-20:]
	}

	return botReply, nil
}

// 每个用户都有自己的Bot上下文
// 用户向机器人发送消息 -> 后端处理消息-> 后端将消息发送给DeepSeek API -> DeepSeek API返回回复 -> 后端将回复发送给用户
// 用户每次聊天时都需要将上文拼起来，作为上下文发送给DeepSeek API

// 全局Bot管理器
var botManager = make(map[string]*Bot)

// 创建新的Bot实例
func NewBot(username string, uuid string, userID int64) *Bot {
	return &Bot{
		Username: username,
		UUID:     uuid,
		UserID:   userID,
		Context:  make([]string, 0),
	}
}

// 获取或创建用户的Bot实例
func GetUserBot(username string, uuid string, userID int64) *Bot {
	if bot, exists := botManager[uuid]; exists {
		return bot
	}
	bot := NewBot(username, uuid, userID)
	botManager[uuid] = bot
	return bot
}

// 清理用户的Bot上下文
func ClearUserBotContext(uuid string) {
	if bot, exists := botManager[uuid]; exists {
		bot.Context = make([]string, 0)
	}
}

// 删除用户的Bot实例
func RemoveUserBot(uuid string) {
	delete(botManager, uuid)
}

func InitBot() {
	// 初始化机器人管理器
	// 每个用户都有自己独立的Bot实例和上下文
	// Bot上下文包含用户的历史对话记录，用于维持连续的对话体验
	botManager = make(map[string]*Bot)
}
