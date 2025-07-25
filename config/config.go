package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	MySQL struct {
		DSN string `mapstructure:"dsn"`
	}
	JWT struct {
		ApiSecret string `mapstructure:"api_secret"`
	}
	Server struct {
		Port int    `mapstructure:"port"`
		Host string `mapstructure:"host"`
	}
}

var AppConfig Config

func LoadConfig() Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../../config")
	viper.AddConfigPath("config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败，错误信息：%v", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("解析配置文件失败，错误信息：%v", err)
	}
	return AppConfig
}
