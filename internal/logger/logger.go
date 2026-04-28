// Package logger 提供基于 zap 和 lumberjack 的结构化日志功能。
package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Config 包含用于日志文件轮转的配置。
type Config struct {
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	Level      zapcore.Level
	Console    bool // 是否同时输出到控制台
}

// Init 使用给定配置初始化日志记录器。
//
//	Init(cfg Config) (*zap.Logger, error)
//
// 创建使用 ISO8601 时间戳的 JSON 编码器和 lumberjack 轮转器。
// 如果 cfg.Console 为 true，则同时输出到控制台（开发调试模式）。
// 返回日志实例和遇到的任何错误。
func Init(cfg Config) (*zap.Logger, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}),
		cfg.Level,
	)

	var core zapcore.Core
	if cfg.Console {
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), cfg.Level)
		core = zapcore.NewTee(fileCore, consoleCore)
	} else {
		core = fileCore
	}

	return zap.New(core, zap.AddCaller()), nil
}

// MustInit 初始化日志记录器，失败时 panic。
//
//	MustInit(cfg Config) *zap.Logger
//
// 用于初始化失败是致命的应用程序。
// 如果初始化失败，则使用格式化错误消息 panic。
func MustInit(cfg Config) *zap.Logger {
	logger, err := Init(cfg)
	if err != nil {
		panic(fmt.Sprintf("初始化日志记录器失败: %v", err))
	}
	return logger
}
