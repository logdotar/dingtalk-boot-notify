// DingTalk Boot Notify 是一个在开机时发送钉钉通知的 Windows 服务程序。
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"dingtalk-boot-notify/internal/config"
	"dingtalk-boot-notify/internal/logger"
	"dingtalk-boot-notify/internal/worker"

	"github.com/kardianos/service"
	"go.uber.org/zap"
)

// 版本信息，由构建脚本注入
var (
	Version     = "dev"
	BuildTime   = "unknown"
	BuildCommit = "unknown"
)

var log *zap.Logger

// program 实现 kardianos/service 的 service.Service 接口。
type program struct {
	worker *worker.Worker
	ctx    context.Context
	cancel context.CancelFunc
}

// Start 在服务启动时调用。
//
//	(p *program) Start(s service.Service) error
//
// 它初始化 worker 上下文并启动 worker goroutine。
// 如果上下文创建失败则返回错误。
func (p *program) Start(s service.Service) error {
	log.Info("服务正在启动...",
		zap.String("版本", Version),
		zap.String("构建时间", BuildTime),
		zap.String("构建提交", BuildCommit),
	)
	p.ctx, p.cancel = context.WithCancel(context.Background())
	go p.run()
	return nil
}

// run 执行 worker 并记录其结果。
//
//	(p *program) run()
//
// 阻塞直到 worker 完成，然后等待上下文取消。
func (p *program) run() {
	err := p.worker.Run(p.ctx)
	if err != nil {
		log.Error("Worker 执行出错", zap.Error(err))
	} else {
		log.Info("Worker 执行完成。")
	}
	<-p.ctx.Done()
}

// Stop 在服务停止时调用。
//
//	(p *program) Stop(s service.Service) error
//
// 它取消 worker 上下文以发出关闭信号。
// 如果取消失败则返回错误。
func (p *program) Stop(s service.Service) error {
	log.Info("服务正在停止...")
	p.cancel()
	return nil
}

// main 是应用程序入口点。
//
//	main()
//
// 其它:
//  1. 确定可执行文件目录
//  2. 从 config.json 加载配置
//  3. 初始化日志记录器
//  4. 创建并启动 Windows 服务
//  5. 设置信号处理器以实现优雅关闭
//
//go:generate go-winres make --arch amd64 --product-version git-tag --file-version git-tag
func main() {
	execPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取可执行文件路径失败: %v\n", err)
		os.Exit(1)
	}
	workDir := filepath.Dir(execPath)

	configPath := filepath.Join(workDir, "config.json")

	logCfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	logLevel, err := logCfg.Log.ToZapLevel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "无效的日志级别: %v\n", err)
		os.Exit(1)
	}

	log = logger.MustInit(logger.Config{
		Filename:   filepath.Join(workDir, logCfg.Log.Filename),
		MaxSize:    logCfg.Log.MaxSize,
		MaxBackups: logCfg.Log.MaxBackups,
		MaxAge:     logCfg.Log.MaxAge,
		Compress:   logCfg.Log.Compress,
		Level:      logLevel,
		Console:    logCfg.Log.Console,
	})
	defer func() { _ = log.Sync() }()

	log.Info("正在加载配置...", zap.String("路径", configPath))

	svcConfig := &service.Config{
		Name:        "DingTalkBootNotify",
		DisplayName: "DingTalk Boot Notify",
		Description: "电脑开机后自动发送钉钉机器人消息提醒",
		UserName:    "NT AUTHORITY\\LocalService",
	}

	prg := &program{
		worker: worker.NewWorker(logCfg, log),
	}

	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal("创建服务失败", zap.Error(err))
	}

	if len(os.Args) > 1 {
		fs := flag.NewFlagSet("svc", flag.ContinueOnError)
		if err := fs.Parse(os.Args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "解析命令行参数失败: %v\n", err)
			os.Exit(1)
		}
		args := fs.Args()
		if len(args) > 0 {
			err = service.Control(s, args[0])
			if err != nil {
				log.Fatal("服务控制失败", zap.Error(err))
			}
			return
		}
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Info("收到关闭信号")
		if err := s.Stop(); err != nil {
			log.Error("停止服务失败", zap.Error(err))
		}
	}()

	log.Info("正在运行服务...")
	err = s.Run()
	if err != nil {
		log.Fatal("服务运行失败", zap.Error(err))
	}
}
