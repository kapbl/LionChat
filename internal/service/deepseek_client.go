package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"cchat/pkg/logger"

	"go.uber.org/zap"
)

// DeepSeek API 配置
type DeepSeekConfig struct {
	APIKey      string  `yaml:"api_key"`
	BaseURL     string  `yaml:"base_url"`
	Model       string  `yaml:"model"`
	MaxTokens   int     `yaml:"max_tokens"`
	Temperature float64 `yaml:"temperature"`
	Timeout     int     `yaml:"timeout"`
}

// DeepSeek API 消息结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeek API 请求结构
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// DeepSeek API 响应结构
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// DeepSeek API 错误响应
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// DeepSeek 客户端
type DeepSeekClient struct {
	config     DeepSeekConfig
	httpClient *http.Client
}

// 创建新的 DeepSeek 客户端
func NewDeepSeekClient(config DeepSeekConfig) *DeepSeekClient {
	return &DeepSeekClient{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   10 * time.Second,
					KeepAlive: 10 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 30 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
				IdleConnTimeout:       90 * time.Second,
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   10,
			},
		},
	}
}

// 调用 DeepSeek Chat Completion API
func (c *DeepSeekClient) ChatCompletion(ctx context.Context, messages []Message) (*ChatResponse, error) {
	// 构建请求体
	request := ChatRequest{
		Model:       c.config.Model,
		Messages:    messages,
		Stream:      false,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
	}

	// 序列化请求体
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 重试机制
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// 指数退避重试
			waitTime := time.Duration(attempt*attempt) * time.Second
			logger.Info("DeepSeek API重试",
				zap.Int("attempt", attempt+1),
				zap.Duration("wait_time", waitTime))
			time.Sleep(waitTime)
		}

		// 创建带超时的上下文
		reqCtx, cancel := context.WithTimeout(ctx, time.Duration(c.config.Timeout)*time.Second)
		defer cancel()

		// 创建HTTP请求
		url := c.config.BaseURL + "/chat/completions"
		req, err := http.NewRequestWithContext(reqCtx, "POST", url, bytes.NewBuffer(requestBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// 设置请求头
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
		req.Header.Set("User-Agent", "LionChat/1.0")

		// 发送请求
		resp, err := c.httpClient.Do(req)
		if err != nil {
			// 检查是否是超时错误
			if isTimeoutError(err) && attempt < maxRetries-1 {
				logger.Warn("DeepSeek API请求超时，准备重试",
					zap.Error(err),
					zap.Int("attempt", attempt+1))
				continue
			}
			return nil, fmt.Errorf("failed to send request: %w", err)
		}

		// 读取响应体（使用带缓冲的读取）
		respBody, err := c.readResponseWithTimeout(resp, reqCtx)
		resp.Body.Close()
		if err != nil {
			if isTimeoutError(err) && attempt < maxRetries-1 {
				logger.Warn("DeepSeek API响应读取超时，准备重试",
					zap.Error(err),
					zap.Int("attempt", attempt+1))
				continue
			}
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		// 检查HTTP状态码
		if resp.StatusCode != http.StatusOK {
			var errorResp ErrorResponse
			if err := json.Unmarshal(respBody, &errorResp); err != nil {
				return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
			}
			// 对于某些错误码进行重试
			if (resp.StatusCode == 429 || resp.StatusCode >= 500) && attempt < maxRetries-1 {
				logger.Warn("DeepSeek API返回可重试错误",
					zap.Int("status_code", resp.StatusCode),
					zap.String("error", errorResp.Error.Message),
					zap.Int("attempt", attempt+1))
				continue
			}
			return nil, fmt.Errorf("API request failed: %s", errorResp.Error.Message)
		}

		// 解析响应
		var chatResp ChatResponse
		if err := json.Unmarshal(respBody, &chatResp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}

		logger.Info("DeepSeek API调用成功",
			zap.String("model", chatResp.Model),
			zap.Int("total_tokens", chatResp.Usage.TotalTokens),
			zap.Int("attempt", attempt+1))

		return &chatResp, nil
	}

	return nil, fmt.Errorf("DeepSeek API调用失败，已达到最大重试次数")
}

// 检查是否是超时错误
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	// 检查网络超时
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}
	// 检查上下文超时
	if err == context.DeadlineExceeded {
		return true
	}
	// 检查错误消息中的超时关键词
	errorMsg := strings.ToLower(err.Error())
	return strings.Contains(errorMsg, "timeout") ||
		strings.Contains(errorMsg, "deadline exceeded") ||
		strings.Contains(errorMsg, "context canceled")
}

// 带超时的响应读取
func (c *DeepSeekClient) readResponseWithTimeout(resp *http.Response, ctx context.Context) ([]byte, error) {
	// 创建一个带缓冲的channel来接收结果
	type result struct {
		data []byte
		err  error
	}
	resultChan := make(chan result, 1)

	// 在goroutine中读取响应
	go func() {
		data, err := io.ReadAll(resp.Body)
		resultChan <- result{data: data, err: err}
	}()

	// 等待结果或超时
	select {
	case res := <-resultChan:
		return res.data, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// 获取默认配置
func GetDefaultDeepSeekConfig() DeepSeekConfig {
	return DeepSeekConfig{
		APIKey:      "", // 需要从环境变量或配置文件中获取
		BaseURL:     "https://api.deepseek.com",
		Model:       "deepseek-chat",
		MaxTokens:   4000,
		Temperature: 0.7,
		Timeout:     30,
	}
}
