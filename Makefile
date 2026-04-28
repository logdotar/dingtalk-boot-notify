# Makefile for DingTalk Boot Notify

# 构建目标
TARGET = dingtalk-boot-notify

# Go 命令
GO = go

# 版本信息
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "unknown")
BUILD_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")

# 构建参数
CGO_ENABLED = 0
LDFLAGS = -ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.BuildCommit=$(BUILD_COMMIT)"
BUILD_FLAGS = $(LDFLAGS) -trimpath

# 测试参数
TEST_FLAGS = -v -race

# 构建目录
OUTPUT_DIR = .

# 可执行文件路径
EXE = $(OUTPUT_DIR)/$(TARGET).exe

# 默认目标
.PHONY: all
all: build

# 构建可执行文件
.PHONY: build
build:
	@echo "正在构建..."
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "CGO_ENABLED: $(CGO_ENABLED)"
	@echo "使用 trimpath: 是"
	$(GO) build $(BUILD_FLAGS) -o $(EXE) main.go
	@echo "构建完成: $(EXE)"

# 构建调试版本
.PHONY: build-debug
build-debug:
	@echo "正在构建调试版本..."
	$(GO) build -gcflags="all=-N -l" -trimpath -o $(EXE) main.go
	@echo "调试版本构建完成: $(EXE)"

# 运行测试
.PHONY: test
test:
	@echo "运行测试..."
	$(GO) test $(TEST_FLAGS) ./internal/...

# 运行代码检查
.PHONY: vet
vet:
	@echo "运行代码检查..."
	$(GO) vet ./...

# 清理依赖并更新
.PHONY: tidy
tidy:
	@echo "更新依赖..."
	$(GO) mod tidy

# 清理构建产物
.PHONY: clean
clean:
	@echo "清理构建产物..."
	@if exist "$(EXE)" del "$(EXE)" >nul 2>&1 || true
	@if exist "service.log" del "service.log" >nul 2>&1 || true
	@if exist "service.log.*" del "service.log.*" >nul 2>&1 || true
	@if exist "vendor" rmdir /s /q "vendor" >nul 2>&1 || true
	@echo "清理完成"

# 运行程序（用于测试）
.PHONY: run
run:
	@echo "运行程序..."
	$(GO) run main.go

# 安装服务
.PHONY: install
install:
	@if not exist "$(EXE)" (
		@echo "错误: 可执行文件不存在，请先运行 make build"
		@exit 1
	)
	@echo "安装服务..."
	$(EXE) install
	@echo "服务安装完成"

# 卸载服务
.PHONY: uninstall
uninstall:
	@echo "卸载服务..."
	$(EXE) uninstall
	@echo "服务卸载完成"

# 启动服务
.PHONY: start
start:
	@echo "启动服务..."
	$(EXE) start
	@echo "服务已启动"

# 停止服务
.PHONY: stop
stop:
	@echo "停止服务..."
	$(EXE) stop
	@echo "服务已停止"

# 重启服务
.PHONY: restart
restart:
	@echo "重启服务..."
	$(EXE) restart
	@echo "服务已重启"

# 查看服务状态
.PHONY: status
status:
	@echo "查看服务状态..."
	$(EXE) status

# 交叉编译
.PHONY: cross-compile
cross-compile:
	@echo "开始交叉编译..."
	@set "GOOS=linux" && set "GOARCH=amd64" && set "CGO_ENABLED=0" && $(GO) build $(LDFLAGS) -trimpath -o $(TARGET)-linux-amd64
	@set "GOOS=darwin" && set "GOARCH=amd64" && set "CGO_ENABLED=0" && $(GO) build $(LDFLAGS) -trimpath -o $(TARGET)-darwin-amd64
	@echo "交叉编译完成"

# 查看帮助
.PHONY: help
help:
	@echo "DingTalk Boot Notify 构建脚本"
	@echo ""
	@echo "可用命令:"
	@echo "  make build         - 构建可执行文件"
	@echo "  make build-debug   - 构建调试版本"
	@echo "  make test          - 运行测试（带竞态检测）"
	@echo "  make vet           - 运行代码检查"
	@echo "  make tidy          - 清理并更新依赖"
	@echo "  make clean         - 清理构建产物"
	@echo "  make run           - 运行程序（用于测试）"
	@echo "  make install       - 安装服务"
	@echo "  make uninstall     - 卸载服务"
	@echo "  make start         - 启动服务"
	@echo "  make stop          - 停止服务"
	@echo "  make restart       - 重启服务"
	@echo "  make status        - 查看服务状态"
	@echo "  make cross-compile - 交叉编译多平台版本"
	@echo "  make help          - 查看帮助"
