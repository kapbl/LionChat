package config

import (
	"cchat/pkg/logger"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Config 应用配置结构
type Config struct {
	MySQL struct {
		DSN             string `mapstructure:"dsn"`
		MaxIdleConns    int    `mapstructure:"max_idle_conns" default:"10"`
		MaxOpenConns    int    `mapstructure:"max_open_conns" default:"100"`
		ConnMaxLifetime int    `mapstructure:"conn_max_lifetime" default:"3600"` // 秒
	}
	JWT struct {
		ApiSecret   string `mapstructure:"api_secret"`
		ExpiresTime int    `mapstructure:"expires_time" default:"86400"` // 秒
		Issuer      string `mapstructure:"issuer" default:"chat-lion"`
		RefreshTime int    `mapstructure:"refresh_time" default:"43200"` // 秒
	}
	Server struct {
		Port         int    `mapstructure:"port" default:"8081"`
		Host         string `mapstructure:"host" default:"127.0.0.1"`
		Environment  string `mapstructure:"environment" default:"development"`
		ServiceName  string `mapstructure:"service_name" default:"chat-lion"`
		ReadTimeout  int    `mapstructure:"read_timeout" default:"60"`  // 秒
		WriteTimeout int    `mapstructure:"write_timeout" default:"60"` // 秒
	}
	Redis struct {
		Addr         string `mapstructure:"addr" default:"localhost:6379"`
		Password     string `mapstructure:"password"`
		DB           int    `mapstructure:"db" default:"0"`
		PoolSize     int    `mapstructure:"pool_size" default:"100"`
		MinIdleConns int    `mapstructure:"min_idle_conns" default:"10"`
	}
	Log struct {
		Level      string `mapstructure:"level" default:"info"`
		FilePath   string `mapstructure:"file_path" default:"logs"`
		MaxSize    int    `mapstructure:"max_size" default:"100"`   // MB
		MaxBackups int    `mapstructure:"max_backups" default:"10"` // 保留的旧文件个数
		MaxAge     int    `mapstructure:"max_age" default:"30"`     // 天
		Compress   bool   `mapstructure:"compress" default:"true"`
	}
	Kafka struct {
		Brokers []string `mapstructure:"brokers"`
		Topics  struct {
			ChatMessages  string `mapstructure:"chat_messages" default:"chat-messages"`
			UserEvents    string `mapstructure:"user_events" default:"user-events"`
			GroupMessages string `mapstructure:"group_messages" default:"group-messages"`
		} `mapstructure:"topics"`
		Consumer struct {
			GroupID         string `mapstructure:"group_id" default:"chat-consumer-group"`
			AutoOffsetReset string `mapstructure:"auto_offset_reset" default:"latest"`
		} `mapstructure:"consumer"`
		Producer struct {
			Acks         string `mapstructure:"acks" default:"all"`
			Retries      int    `mapstructure:"retries" default:"3"`
			BatchSize    int    `mapstructure:"batch_size" default:"16384"`
			LingerMs     int    `mapstructure:"linger_ms" default:"1"`
			BufferMemory int    `mapstructure:"buffer_memory" default:"33554432"`
		} `mapstructure:"producer"`
	}
}

var (
	// AppConfig 全局配置对象
	AppConfig Config
	// configPaths 配置文件搜索路径
	configPaths = []string{
		"config/",
		"../config/",
		"../../config/",
	}
)

// LoadConfig 加载配置
func LoadConfig() Config {
	// 设置环境变量前缀
	viper.SetEnvPrefix("CHAT")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 设置默认配置文件名和类型
	viper.SetConfigName("config.instance2")
	viper.SetConfigType("yaml")

	// 添加配置文件搜索路径
	for _, path := range configPaths {
		viper.AddConfigPath(path)
	}
	logger.Info("配置文件搜索路径", zap.Strings("paths", configPaths))
	//"[b:\\projects\\golang\\chatLion\\cmd\\app\\config
	// b:\\projects\\golang\\chatLion\\cmd\\config
	// b:\\projects\\golang\\chatLion\\config]"
	// 读取环境变量中的配置文件路径
	if configPath := os.Getenv("CHAT_CONFIG_PATH"); configPath != "" {
		viper.AddConfigPath(configPath)
	}

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		logger.Error("读取配置文件失败", zap.Error(err))
		// 如果是配置文件不存在，则创建默认配置文件
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := createDefaultConfig(); err != nil {
				logger.Fatal("创建默认配置文件失败", zap.Error(err))
			}
		} else {
			logger.Fatal("配置文件错误", zap.Error(err))
		}
	}

	// 解析配置到结构体
	if err := viper.Unmarshal(&AppConfig); err != nil {
		logger.Fatal("解析配置文件失败", zap.Error(err))
	}

	// 打印当前使用的配置文件路径
	logger.Info("加载配置文件成功",
		zap.String("config_file", viper.ConfigFileUsed()),
		zap.String("environment", AppConfig.Server.Environment))

	return AppConfig
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig() error {
	// 确保配置目录存在
	configDir := "config"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 创建默认配置文件
	configPath := filepath.Join(configDir, "config.yaml")
	defaultConfig := `MySQL:
  dsn: "root:password@tcp(127.0.0.1:3306)/chat?charset=utf8mb4&parseTime=True&loc=Local&timeout=1000ms"
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 3600

JWT:
  api_secret: "chat-api-secret"
  expires_time: 86400
  issuer: "chat-lion"
  refresh_time: 43200

Server:
  port: 8081
  host: "127.0.0.1"
  environment: "development"
  service_name: "chat-lion"
  read_timeout: 60
  write_timeout: 60

Redis:
  addr: "localhost:6379"
  password: ""
  db: 0
  pool_size: 100
  min_idle_conns: 10

Log:
  level: "info"
  file_path: "logs"
  max_size: 100
  max_backups: 10
  max_age: 30
  compress: true
`

	// 写入默认配置
	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("写入默认配置文件失败: %w", err)
	}

	// 重新加载配置
	return viper.ReadInConfig()
}
