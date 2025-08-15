package logger

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

// InitLogger 初始化日志
func InitLogger() {
	// 设置日志级别
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zapcore.InfoLevel)

	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// --- 文件输出 ---
	logDir := "logs/server"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(fmt.Sprintf("创建日志目录失败: %v", err))
	}

	logFileName := fmt.Sprintf("%s/%s.log", logDir, time.Now().Format("2006-01-02-15-04-05"))
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("打开日志文件失败: %v", err))
	}

	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	fileWriter := zapcore.AddSync(logFile)
	fileCore := zapcore.NewCore(fileEncoder, fileWriter, atomicLevel)

	// --- 控制台输出 ---
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleOutput := zapcore.Lock(os.Stdout)
	consoleCore := zapcore.NewCore(consoleEncoder, consoleOutput, atomicLevel)

	// 创建核心
	core := zapcore.NewTee(
		consoleCore,
		fileCore,
	)

	// 创建logger
	Logger = zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
	defer Logger.Sync()
}

// Info wrapper for zap.Logger.Info
func Info(msg string, fields ...zapcore.Field) {
	Logger.Info(msg, fields...)
}

// Error wrapper for zap.Logger.Error
func Error(msg string, fields ...zapcore.Field) {
	Logger.Error(msg, fields...)
}

// Fatal wrapper for zap.Logger.Fatal
func Fatal(msg string, fields ...zapcore.Field) {
	Logger.Fatal(msg, fields...)
}

// Debug wrapper for zap.Logger.Debug
func Debug(msg string, fields ...zapcore.Field) {
	Logger.Debug(msg, fields...)
}

// Warn wrapper for zap.Logger.Warn
func Warn(msg string, fields ...zapcore.Field) {
	Logger.Warn(msg, fields...)
}
