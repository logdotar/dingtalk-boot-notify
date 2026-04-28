package logger

import (
	"os"
	"testing"

	"go.uber.org/zap/zapcore"
)

// TestInit_Success 测试日志初始化成功。
func TestInit_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-log-*.log")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	tmpFileName := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpFileName)
		t.Fatalf("关闭临时文件失败: %v", err)
	}

	cfg := Config{
		Filename:   tmpFileName,
		MaxSize:    1,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   false,
		Level:      zapcore.InfoLevel,
		Console:    false,
	}

	logger, err := Init(cfg)
	if err != nil {
		_ = os.Remove(tmpFileName)
		t.Fatalf("期望无错误，得到 %v", err)
	}
	if logger == nil {
		_ = os.Remove(tmpFileName)
		t.Fatal("期望非 nil logger")
	}

	_ = logger.Sync()
	_ = os.Remove(tmpFileName)
}

// TestInit_WithConsole 测试带控制台输出的日志初始化。
func TestInit_WithConsole(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-log-*.log")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	tmpFileName := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpFileName)
		t.Fatalf("关闭临时文件失败: %v", err)
	}

	cfg := Config{
		Filename:   tmpFileName,
		MaxSize:    1,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   false,
		Level:      zapcore.DebugLevel,
		Console:    true,
	}

	logger, err := Init(cfg)
	if err != nil {
		_ = os.Remove(tmpFileName)
		t.Fatalf("期望无错误，得到 %v", err)
	}
	if logger == nil {
		_ = os.Remove(tmpFileName)
		t.Fatal("期望非 nil logger")
	}

	_ = logger.Sync()
	_ = os.Remove(tmpFileName)
}

// TestInit_WithDifferentLevels 测试不同日志级别。
func TestInit_WithDifferentLevels(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-log-*.log")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	tmpFileName := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpFileName)
		t.Fatalf("关闭临时文件失败: %v", err)
	}

	levels := []zapcore.Level{
		zapcore.DebugLevel,
		zapcore.InfoLevel,
		zapcore.WarnLevel,
		zapcore.ErrorLevel,
		zapcore.DPanicLevel,
		zapcore.PanicLevel,
		zapcore.FatalLevel,
	}

	for _, level := range levels {
		cfg := Config{
			Filename:   tmpFileName,
			MaxSize:    1,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   false,
			Level:      level,
			Console:    false,
		}

		logger, err := Init(cfg)
		if err != nil {
			t.Errorf("级别 %v: 期望无错误，得到 %v", level, err)
		}
		if logger == nil {
			t.Errorf("级别 %v: 期望非 nil logger", level)
		}
		if logger != nil {
			_ = logger.Sync()
		}
	}
	_ = os.Remove(tmpFileName)
}

// TestMustInit_Success 测试 MustInit 成功。
func TestMustInit_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-log-*.log")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	tmpFileName := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpFileName)
		t.Fatalf("关闭临时文件失败: %v", err)
	}

	cfg := Config{
		Filename:   tmpFileName,
		MaxSize:    1,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   false,
		Level:      zapcore.InfoLevel,
		Console:    false,
	}

	logger := MustInit(cfg)
	if logger == nil {
		_ = os.Remove(tmpFileName)
		t.Fatal("期望非 nil logger")
	}

	_ = logger.Sync()
	_ = os.Remove(tmpFileName)
}

// TestConfig_Defaults 测试默认配置值。
func TestConfig_Defaults(t *testing.T) {
	cfg := Config{
		Filename:   "test.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     30,
		Compress:   true,
		Level:      zapcore.InfoLevel,
		Console:    false,
	}

	if cfg.Filename != "test.log" {
		t.Errorf("期望 test.log，得到 %s", cfg.Filename)
	}
	if cfg.MaxSize != 10 {
		t.Errorf("期望 10，得到 %d", cfg.MaxSize)
	}
	if cfg.MaxBackups != 3 {
		t.Errorf("期望 3，得到 %d", cfg.MaxBackups)
	}
	if cfg.MaxAge != 30 {
		t.Errorf("期望 30，得到 %d", cfg.MaxAge)
	}
	if !cfg.Compress {
		t.Error("期望 Compress 为 true")
	}
	if cfg.Level != zapcore.InfoLevel {
		t.Errorf("期望 InfoLevel，得到 %v", cfg.Level)
	}
	if cfg.Console {
		t.Error("期望 Console 为 false")
	}
}
