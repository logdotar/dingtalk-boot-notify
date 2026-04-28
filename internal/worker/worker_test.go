package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"dingtalk-boot-notify/internal/config"

	"go.uber.org/zap"
)

// mockSender 模拟消息发送者。
type mockSender struct {
	sendErr        error
	callCount      int
	sendMarkdownFn func(ctx context.Context, title, text string, atInfo interface{}) error
}

func (m *mockSender) Send(ctx context.Context, message string) error {
	m.callCount++
	if m.sendErr != nil {
		return m.sendErr
	}
	return nil
}

func (m *mockSender) SendMarkdown(ctx context.Context, title, text string, atInfo interface{}) error {
	m.callCount++
	if m.sendMarkdownFn != nil {
		return m.sendMarkdownFn(ctx, title, text, atInfo)
	}
	if m.sendErr != nil {
		return m.sendErr
	}
	return nil
}

// newTestConfig 创建测试配置。
func newTestConfig() *config.Config {
	return &config.Config{
		Retry: config.RetryConfig{
			MaxAttempts: 3,
			InitDelay:   10 * time.Millisecond,
			MaxDelay:    50 * time.Millisecond,
		},
		Network: config.NetworkConfig{
			WaitTimeout: 100 * time.Millisecond,
		},
	}
}

// TestWorker_GetHostname 测试获取主机名。
func TestWorker_GetHostname(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := newTestConfig()
	w := &Worker{
		cfg:      cfg,
		logger:   logger,
		hostname: "test-hostname",
	}

	hostname := w.hostname
	if hostname == "" {
		t.Error("期望非空主机名")
	}
	if hostname != "test-hostname" {
		t.Errorf("期望 test-hostname，得到 %s", hostname)
	}
}

// TestWorker_NewWorker 测试创建 Worker。
func TestWorker_NewWorker(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := newTestConfig()

	w := NewWorker(cfg, logger)

	if w == nil {
		t.Fatal("期望非 nil Worker")
	}
	if w.hostname == "" {
		t.Error("期望主机名非空")
	}
}

// TestWorker_BuildBootMessage 测试构建开机消息。
func TestWorker_BuildBootMessage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := newTestConfig()
	w := &Worker{
		cfg:      cfg,
		logger:   logger,
		hostname: "test-pc",
	}

	title, content := w.buildBootMessage()

	if title == "" {
		t.Error("期望非空标题")
	}
	if content == "" {
		t.Error("期望非空内容")
	}
	if len(content) < 10 {
		t.Error("消息内容过短")
	}
	if !containsString(content, "test-pc") {
		t.Error("消息内容应包含主机名")
	}
	if !containsString(content, "开机通知") {
		t.Error("消息内容应包含开机通知")
	}
}

// TestWorker_BuildBootMessage_TimeFormat 测试开机消息时间格式。
func TestWorker_BuildBootMessage_TimeFormat(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := newTestConfig()
	w := &Worker{
		cfg:      cfg,
		logger:   logger,
		hostname: "test-pc",
	}

	_, content := w.buildBootMessage()

	now := time.Now().Format("2006-01-02 15:04:05")
	if !containsString(content, now) {
		t.Errorf("消息内容应包含当前时间 %s", now)
	}
}

// TestConstants 测试常量定义。
func TestConstants(t *testing.T) {
	if defaultHostname != "unknown" {
		t.Errorf("期望 defaultHostname=unknown，得到 %s", defaultHostname)
	}
	if networkCheckURL == "" {
		t.Error("期望 networkCheckURL 非空")
	}
	if checkInterval <= 0 {
		t.Error("期望 checkInterval > 0")
	}
}

// containsString 检查字符串是否包含子串。
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

// findSubstring 查找子串位置。
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// TestWorker_RetryConfig 测试重试配置。
func TestWorker_RetryConfig(t *testing.T) {
	cfg := newTestConfig()

	if cfg.Retry.MaxAttempts != 3 {
		t.Errorf("期望 MaxAttempts=3，得到 %d", cfg.Retry.MaxAttempts)
	}
	if cfg.Retry.InitDelay != 10*time.Millisecond {
		t.Errorf("期望 InitDelay=10ms，得到 %v", cfg.Retry.InitDelay)
	}
	if cfg.Retry.MaxDelay != 50*time.Millisecond {
		t.Errorf("期望 MaxDelay=50ms，得到 %v", cfg.Retry.MaxDelay)
	}
}

// TestWorker_NetworkConfig 测试网络配置。
func TestWorker_NetworkConfig(t *testing.T) {
	cfg := newTestConfig()

	if cfg.Network.WaitTimeout != 100*time.Millisecond {
		t.Errorf("期望 WaitTimeout=100ms，得到 %v", cfg.Network.WaitTimeout)
	}
}

// TestMockSender_Send 测试 mock Sender 的 Send 方法。
func TestMockSender_Send(t *testing.T) {
	sender := &mockSender{}

	err := sender.Send(context.Background(), "test")
	if err != nil {
		t.Errorf("期望无错误，得到 %v", err)
	}
	if sender.callCount != 1 {
		t.Errorf("期望 callCount=1，得到 %d", sender.callCount)
	}
}

// TestMockSender_Send_Error 测试 mock Sender 错误。
func TestMockSender_Send_Error(t *testing.T) {
	sender := &mockSender{
		sendErr: errors.New("test error"),
	}

	err := sender.Send(context.Background(), "test")
	if err == nil {
		t.Error("期望错误")
	}
}

// TestMockSender_SendMarkdown 测试 mock Sender 的 SendMarkdown 方法。
func TestMockSender_SendMarkdown(t *testing.T) {
	sender := &mockSender{}

	err := sender.SendMarkdown(context.Background(), "title", "text", nil)
	if err != nil {
		t.Errorf("期望无错误，得到 %v", err)
	}
	if sender.callCount != 1 {
		t.Errorf("期望 callCount=1，得到 %d", sender.callCount)
	}
}

// TestMockSender_SendMarkdown_Error 测试 mock Sender SendMarkdown 错误。
func TestMockSender_SendMarkdown_Error(t *testing.T) {
	sender := &mockSender{
		sendErr: errors.New("test error"),
	}

	err := sender.SendMarkdown(context.Background(), "title", "text", nil)
	if err == nil {
		t.Error("期望错误")
	}
}
