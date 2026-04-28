// Package config 提供基于 Viper 的配置加载和验证功能。
package config

import (
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

// Config 包含所有应用程序配置节。
type Config struct {
	DingTalk DingTalkConfig `mapstructure:"dingtalk"`
	Retry    RetryConfig    `mapstructure:"retry"`
	Network  NetworkConfig  `mapstructure:"network"`
	Log      LogConfig      `mapstructure:"log"`
}

// DingTalkConfig 包含钉钉 webhook 配置。
type DingTalkConfig struct {
	Webhook string `mapstructure:"webhook"`
	Secret  string `mapstructure:"secret"`
}

// RetryConfig 包含消息发送的重试策略配置。
type RetryConfig struct {
	MaxAttempts int           `mapstructure:"max_attempts"`
	InitDelay   time.Duration `mapstructure:"init_delay"`
	MaxDelay    time.Duration `mapstructure:"max_delay"`
}

// NetworkConfig 包含网络相关配置。
type NetworkConfig struct {
	WaitTimeout time.Duration `mapstructure:"wait_timeout"`
}

// LogConfig 包含 zap 日志配置。
type LogConfig struct {
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
	Level      string `mapstructure:"level"`
	Console    bool   `mapstructure:"console"`
}

const (
	defaultMaxAttempts = 5
	defaultInitDelay   = 2 * time.Second
	defaultMaxDelay    = 30 * time.Second
	defaultWaitTimeout = 120 * time.Second
)

const (
	defaultLogMaxSize    = 10
	defaultLogMaxBackups = 3
	defaultLogMaxAge     = 30
	defaultLogCompress   = true
	defaultLogLevel      = "info"
	defaultLogConsole    = false
)

// setDefaults 为所有配置字段设置默认值。
func setDefaults(v *viper.Viper) {
	v.SetDefault("retry.max_attempts", defaultMaxAttempts)
	v.SetDefault("retry.init_delay", defaultInitDelay.String())
	v.SetDefault("retry.max_delay", defaultMaxDelay.String())
	v.SetDefault("network.wait_timeout", defaultWaitTimeout.String())

	v.SetDefault("log.max_size", defaultLogMaxSize)
	v.SetDefault("log.max_backups", defaultLogMaxBackups)
	v.SetDefault("log.max_age", defaultLogMaxAge)
	v.SetDefault("log.compress", defaultLogCompress)
	v.SetDefault("log.level", defaultLogLevel)
	v.SetDefault("log.console", defaultLogConsole)
}

// Load 从指定的 JSON 文件路径读取配置。
//
//	configPath: JSON 配置文件的绝对或相对路径
//
// 支持使用 DINGTALK_ 前缀的环境变量覆盖配置。
// 如果文件无法读取、解析或包含无效值，则返回错误。
func Load(configPath string) (*Config, error) {
	v := viper.New()

	setDefaults(v)

	v.SetConfigFile(configPath)
	v.SetConfigType("json")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	v.SetEnvPrefix("DINGTALK")
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &cfg, nil
}

// Validate 检查配置值是否有效。
//
//	(c *Config) Validate() error
//
// 如果任何必填字段缺失或值无效，则返回错误。
func (c *Config) Validate() error {
	if c.DingTalk.Webhook == "" {
		return fmt.Errorf("钉钉 webhook 不能为空")
	}

	if _, err := url.ParseRequestURI(c.DingTalk.Webhook); err != nil {
		return fmt.Errorf("钉钉 webhook 不是有效的 URL: %w", err)
	}

	if c.Retry.MaxAttempts <= 0 {
		return fmt.Errorf("重试次数必须大于 0")
	}
	if c.Retry.InitDelay <= 0 {
		return fmt.Errorf("初始重试延迟必须大于 0")
	}
	if c.Retry.MaxDelay < c.Retry.InitDelay {
		return fmt.Errorf("最大重试延迟必须大于等于初始重试延迟")
	}
	if c.Network.WaitTimeout <= 0 {
		return fmt.Errorf("网络等待超时必须大于 0")
	}
	if c.Log.Filename == "" {
		return fmt.Errorf("日志文件名不能为空")
	}
	if c.Log.MaxSize <= 0 {
		return fmt.Errorf("日志文件最大大小必须大于 0")
	}
	if c.Log.MaxBackups <= 0 {
		return fmt.Errorf("日志文件保留数量必须大于 0")
	}
	if c.Log.MaxAge <= 0 {
		return fmt.Errorf("日志文件保留天数必须大于 0")
	}
	return nil
}

// ToZapLevel 将日志级别字符串转换为 zapcore.Level。
//
//	(c *LogConfig) ToZapLevel() (zapcore.Level, error)
//
// 支持: "debug", "info", "warn", "error", "dpanic", "panic", "fatal"
// 如果级别字符串无效则返回错误。
func (c *LogConfig) ToZapLevel() (zapcore.Level, error) {
	return zapcore.ParseLevel(c.Level)
}
