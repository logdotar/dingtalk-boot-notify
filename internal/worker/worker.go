// Package worker 实现发送钉钉通知的核心业务逻辑。
package worker

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"dingtalk-boot-notify/internal/config"
	"dingtalk-boot-notify/internal/dingtalk"

	"go.uber.org/zap"
)

const (
	defaultHostname = "unknown"
	networkCheckURL = "https://oapi.dingtalk.com"
	checkInterval   = 2 * time.Second
)

// Sender 是发送消息的接口，用于在测试中 mock 钉钉客户端。
type Sender interface {
	Send(ctx context.Context, message string) error
	SendMarkdown(ctx context.Context, title, text string, atInfo *dingtalk.AtInfo) error
}

// Worker 处理发送通知的业务逻辑。
// 它等待网络连接，然后通过配置的 Sender 发送消息。
type Worker struct {
	cfg      *config.Config
	client   *dingtalk.Client
	logger   *zap.Logger
	hostname string
}

// NewWorker 创建 Worker 实例。
//
//	cfg:    应用程序配置
//	logger: zap 日志实例
//
// 返回配置好的 Worker 指针。
func NewWorker(cfg *config.Config, logger *zap.Logger) *Worker {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = defaultHostname
		logger.Warn("获取主机名失败，使用默认值", zap.String("hostname", hostname))
	}

	return &Worker{
		cfg:      cfg,
		client:   dingtalk.NewClient(cfg.DingTalk.Webhook, cfg.DingTalk.Secret),
		logger:   logger,
		hostname: hostname,
	}
}

// Run 执行主要工作流程：等待网络，然后发送开机通知。
//
//	(w *Worker) Run(ctx context.Context) error
//
// 流程：
//  1. 通过 waitForNetwork() 等待网络连接
//  2. 通过 sendWithRetry() 发送 Markdown 格式开机通知
//
// 如果网络等待失败、所有重试耗尽或上下文被取消，则返回错误。
func (w *Worker) Run(ctx context.Context) error {
	if err := w.waitForNetwork(ctx); err != nil {
		return fmt.Errorf("网络等待中断: %w", err)
	}

	return w.sendWithRetry(ctx)
}

// waitForNetwork 等待网络连接可用，通过定期检查钉钉 API 端点来实现。
//
//	(w *Worker) waitForNetwork(ctx context.Context) error
//
// 使用 2 秒间隔的 ticker 检测网络。
// 尊重上下文取消。超时或上下文取消时返回错误。
func (w *Worker) waitForNetwork(ctx context.Context) error {
	w.logger.Info("等待网络连接...")

	client := &http.Client{Timeout: 3 * time.Second}
	deadline := time.Now().Add(w.cfg.Network.WaitTimeout)
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			resp, err := client.Head(networkCheckURL)
			if err == nil {
				if closeErr := resp.Body.Close(); closeErr != nil {
					w.logger.Warn("关闭响应体失败", zap.Error(closeErr))
				}
				w.logger.Info("网络已就绪。")
				return nil
			}
			w.logger.Warn("网络未就绪，重试中...", zap.Error(err))

			if time.Now().After(deadline) {
				return fmt.Errorf("网络超时，等待 %v 后仍无法连接", w.cfg.Network.WaitTimeout)
			}
		}
	}
}

// buildBootMessage 构建开机通知消息（Markdown 格式）。
func (w *Worker) buildBootMessage() (title, content string) {
	title = "【开机提醒】"
	content = fmt.Sprintf(`### 🖥️ 电脑开机通知

**主机名：** %s

> 电脑已成功开机并进入系统

---
⏰ 时间：%s

---
📌 请确认是否为预期操作`,
		w.hostname,
		time.Now().Format(time.DateTime),
	)
	return
}

// sendWithRetry 使用指数退避重试发送消息。
//
//	(w *Worker) sendWithRetry(ctx context.Context) error
//
// 最多尝试 cfg.Retry.MaxAttempts 次，每次重试之间的延迟翻倍
// (init_delay -> 2*init_delay -> ... -> max_delay)。
// 所有重试耗尽或上下文取消时返回错误。
func (w *Worker) sendWithRetry(ctx context.Context) error {
	delay := w.cfg.Retry.InitDelay
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	title, content := w.buildBootMessage()

	for attempt := 1; attempt <= w.cfg.Retry.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			w.logger.Info("发送被关闭信号取消。")
			return ctx.Err()
		case <-ticker.C:
			w.logger.Info("正在发送消息...", zap.Int("尝试次数", attempt))
			err := w.client.SendMarkdown(ctx, title, content, nil)
			if err == nil {
				w.logger.Info("消息发送成功！")
				return nil
			}

			w.logger.Error("消息发送失败",
				zap.Int("尝试次数", attempt),
				zap.Error(err),
				zap.Duration("下次重试等待", delay),
			)

			if attempt == w.cfg.Retry.MaxAttempts {
				return fmt.Errorf("经过 %d 次尝试后仍然失败，最后错误: %w", w.cfg.Retry.MaxAttempts, err)
			}

			delay *= 2
			if delay > w.cfg.Retry.MaxDelay {
				delay = w.cfg.Retry.MaxDelay
			}
			ticker.Reset(delay)
		}
	}
	return nil
}
