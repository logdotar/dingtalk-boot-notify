package config

import (
	"os"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

// TestLoad_Success 测试加载配置成功。
func TestLoad_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-config-*.json")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	tmpFileName := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpFileName)
	}()

	configContent := `{
		"dingtalk": {
			"webhook": "https://oapi.dingtalk.com/robot/send?access_token=test",
			"secret": "test-secret"
		},
		"retry": {
			"max_attempts": 5,
			"init_delay": "2s",
			"max_delay": "30s"
		},
		"network": {
			"wait_timeout": "120s"
		},
		"log": {
			"filename": "service.log",
			"max_size": 10,
			"max_backups": 3,
			"max_age": 30,
			"compress": true,
			"level": "info",
			"console": true
		}
	}`

	if _, err := tmpFile.WriteString(configContent); err != nil {
		_ = tmpFile.Close()
		t.Fatalf("写入配置失败: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("关闭临时文件失败: %v", err)
	}

	cfg, err := Load(tmpFileName)
	if err != nil {
		t.Fatalf("期望无错误，得到 %v", err)
	}

	if cfg.DingTalk.Webhook == "" {
		t.Error("期望 webhook 非空")
	}
	if cfg.Retry.MaxAttempts != 5 {
		t.Errorf("期望 max_attempts=5，得到 %d", cfg.Retry.MaxAttempts)
	}
	if cfg.Log.Console != true {
		t.Error("期望 console=true")
	}
}

// TestLoad_FileNotFound 测试文件不存在。
func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("不存在文件.json")
	if err == nil {
		t.Error("期望错误，因为文件不存在")
	}
}

// TestLoad_InvalidJSON 测试无效 JSON。
func TestLoad_InvalidJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-config-*.json")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	tmpFileName := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpFileName)
	}()

	if _, err := tmpFile.WriteString("invalid json"); err != nil {
		_ = tmpFile.Close()
		t.Fatalf("写入配置失败: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("关闭临时文件失败: %v", err)
	}

	_, err = Load(tmpFileName)
	if err == nil {
		t.Error("期望错误，因为 JSON 无效")
	}
}

// TestLogConfig_ToZapLevel 测试日志级别转换。
func TestLogConfig_ToZapLevel(t *testing.T) {
	tests := []struct {
		level    string
		expected zapcore.Level
		wantErr  bool
	}{
		{"debug", zapcore.DebugLevel, false},
		{"info", zapcore.InfoLevel, false},
		{"warn", zapcore.WarnLevel, false},
		{"error", zapcore.ErrorLevel, false},
		{"dpanic", zapcore.DPanicLevel, false},
		{"panic", zapcore.PanicLevel, false},
		{"fatal", zapcore.FatalLevel, false},
		{"invalid", zapcore.InfoLevel, true},
		{"", zapcore.InfoLevel, false},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			cfg := &LogConfig{Level: tt.level}
			level, err := cfg.ToZapLevel()
			if (err != nil) != tt.wantErr {
				t.Errorf("ToZapLevel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && level != tt.expected {
				t.Errorf("ToZapLevel() = %v, expected %v", level, tt.expected)
			}
		})
	}
}

// TestConfig_Validate_Success 测试验证成功。
func TestConfig_Validate_Success(t *testing.T) {
	cfg := &Config{
		DingTalk: DingTalkConfig{
			Webhook: "https://oapi.dingtalk.com/robot/send?access_token=test",
			Secret:  "test-secret",
		},
		Retry: RetryConfig{
			MaxAttempts: 5,
			InitDelay:   2 * time.Second,
			MaxDelay:    30 * time.Second,
		},
		Network: NetworkConfig{
			WaitTimeout: 120 * time.Second,
		},
		Log: LogConfig{
			Filename:   "test.log",
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     30,
			Compress:   true,
			Level:      "info",
		},
	}

	err := cfg.Validate()
	if err != nil {
		t.Fatalf("期望无错误，得到 %v", err)
	}
}

// TestConfig_Validate_EmptyWebhook 测试空 webhook。
func TestConfig_Validate_EmptyWebhook(t *testing.T) {
	cfg := &Config{
		DingTalk: DingTalkConfig{
			Webhook: "",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("期望返回错误，因为 webhook 为空")
	}
}

// TestConfig_Validate_InvalidURL 测试无效 URL。
func TestConfig_Validate_InvalidURL(t *testing.T) {
	cfg := &Config{
		DingTalk: DingTalkConfig{
			Webhook: "not-a-valid-url",
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("期望返回错误，因为 URL 无效")
	}
}

// TestConfig_Validate_RetryAttempts 测试重试次数验证。
func TestConfig_Validate_RetryAttempts(t *testing.T) {
	cfg := &Config{
		DingTalk: DingTalkConfig{
			Webhook: "https://oapi.dingtalk.com/robot/send?access_token=test",
		},
		Retry: RetryConfig{
			MaxAttempts: 0,
			InitDelay:   2 * time.Second,
			MaxDelay:    30 * time.Second,
		},
		Network: NetworkConfig{
			WaitTimeout: 120 * time.Second,
		},
		Log: LogConfig{
			Filename:   "test.log",
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     30,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("期望返回错误，因为 max_attempts=0")
	}
}

// TestConfig_Validate_RetryDelays 测试重试延迟验证。
func TestConfig_Validate_RetryDelays(t *testing.T) {
	cfg := &Config{
		DingTalk: DingTalkConfig{
			Webhook: "https://oapi.dingtalk.com/robot/send?access_token=test",
		},
		Retry: RetryConfig{
			MaxAttempts: 5,
			InitDelay:   10 * time.Second,
			MaxDelay:    5 * time.Second,
		},
		Network: NetworkConfig{
			WaitTimeout: 120 * time.Second,
		},
		Log: LogConfig{
			Filename:   "test.log",
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     30,
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("期望返回错误，因为 max_delay < init_delay")
	}
}

// TestConfig_Validate_LogConfig 测试日志配置验证。
func TestConfig_Validate_LogConfig(t *testing.T) {
	tests := []struct {
		name    string
		logCfg  LogConfig
		wantErr bool
	}{
		{"文件名为空", LogConfig{Filename: "", MaxSize: 10, MaxBackups: 3, MaxAge: 30}, true},
		{"max_size为零", LogConfig{Filename: "test.log", MaxSize: 0, MaxBackups: 3, MaxAge: 30}, true},
		{"max_backups为零", LogConfig{Filename: "test.log", MaxSize: 10, MaxBackups: 0, MaxAge: 30}, true},
		{"max_age为零", LogConfig{Filename: "test.log", MaxSize: 10, MaxBackups: 3, MaxAge: 0}, true},
		{"有效配置", LogConfig{Filename: "test.log", MaxSize: 10, MaxBackups: 3, MaxAge: 30}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				DingTalk: DingTalkConfig{
					Webhook: "https://oapi.dingtalk.com/robot/send?access_token=test",
				},
				Retry: RetryConfig{
					MaxAttempts: 5,
					InitDelay:   2 * time.Second,
					MaxDelay:    30 * time.Second,
				},
				Network: NetworkConfig{
					WaitTimeout: 120 * time.Second,
				},
				Log: tt.logCfg,
			}

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDingTalkConfig_Structure 测试钉钉配置结构。
func TestDingTalkConfig_Structure(t *testing.T) {
	cfg := DingTalkConfig{
		Webhook: "https://oapi.dingtalk.com/robot/send?access_token=test",
		Secret:  "test-secret",
	}

	if cfg.Webhook == "" {
		t.Error("期望 webhook 非空")
	}
	if cfg.Secret == "" {
		t.Error("期望 secret 非空")
	}
}

// TestRetryConfig_Structure 测试重试配置结构。
func TestRetryConfig_Structure(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts: 5,
		InitDelay:   2 * time.Second,
		MaxDelay:    30 * time.Second,
	}

	if cfg.MaxAttempts != 5 {
		t.Errorf("期望 MaxAttempts=5，得到 %d", cfg.MaxAttempts)
	}
	if cfg.InitDelay != 2*time.Second {
		t.Errorf("期望 InitDelay=2s，得到 %v", cfg.InitDelay)
	}
	if cfg.MaxDelay != 30*time.Second {
		t.Errorf("期望 MaxDelay=30s，得到 %v", cfg.MaxDelay)
	}
}

// TestNetworkConfig_Structure 测试网络配置结构。
func TestNetworkConfig_Structure(t *testing.T) {
	cfg := NetworkConfig{
		WaitTimeout: 120 * time.Second,
	}

	if cfg.WaitTimeout != 120*time.Second {
		t.Errorf("期望 WaitTimeout=120s，得到 %v", cfg.WaitTimeout)
	}
}

// TestConstants 测试常量定义。
func TestConstants(t *testing.T) {
	if defaultMaxAttempts != 5 {
		t.Errorf("期望 defaultMaxAttempts=5，得到 %d", defaultMaxAttempts)
	}
	if defaultInitDelay != 2*time.Second {
		t.Errorf("期望 defaultInitDelay=2s，得到 %v", defaultInitDelay)
	}
	if defaultMaxDelay != 30*time.Second {
		t.Errorf("期望 defaultMaxDelay=30s，得到 %v", defaultMaxDelay)
	}
	if defaultWaitTimeout != 120*time.Second {
		t.Errorf("期望 defaultWaitTimeout=120s，得到 %v", defaultWaitTimeout)
	}
	if defaultLogMaxSize != 10 {
		t.Errorf("期望 defaultLogMaxSize=10，得到 %d", defaultLogMaxSize)
	}
	if defaultLogMaxBackups != 3 {
		t.Errorf("期望 defaultLogMaxBackups=3，得到 %d", defaultLogMaxBackups)
	}
	if defaultLogMaxAge != 30 {
		t.Errorf("期望 defaultLogMaxAge=30，得到 %d", defaultLogMaxAge)
	}
	if !defaultLogCompress {
		t.Error("期望 defaultLogCompress=true")
	}
	if defaultLogLevel != "info" {
		t.Errorf("期望 defaultLogLevel=info，得到 %s", defaultLogLevel)
	}
	if defaultLogConsole {
		t.Error("期望 defaultLogConsole=false")
	}
}

// TestConfig_Structure 测试配置结构。
func TestConfig_Structure(t *testing.T) {
	cfg := Config{
		DingTalk: DingTalkConfig{
			Webhook: "https://oapi.dingtalk.com/robot/send?access_token=test",
			Secret:  "test-secret",
		},
		Retry: RetryConfig{
			MaxAttempts: 5,
			InitDelay:   2 * time.Second,
			MaxDelay:    30 * time.Second,
		},
		Network: NetworkConfig{
			WaitTimeout: 120 * time.Second,
		},
		Log: LogConfig{
			Filename:   "test.log",
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     30,
			Compress:   true,
			Level:      "info",
		},
	}

	if cfg.DingTalk.Webhook == "" {
		t.Error("期望 DingTalk.Webhook 非空")
	}
	if cfg.Retry.MaxAttempts != 5 {
		t.Errorf("期望 Retry.MaxAttempts=5，得到 %d", cfg.Retry.MaxAttempts)
	}
	if cfg.Network.WaitTimeout != 120*time.Second {
		t.Errorf("期望 Network.WaitTimeout=120s，得到 %v", cfg.Network.WaitTimeout)
	}
	if cfg.Log.Filename != "test.log" {
		t.Errorf("期望 Log.Filename=test.log，得到 %s", cfg.Log.Filename)
	}
}
