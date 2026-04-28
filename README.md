# DingTalk Boot Notify

电脑开机后自动发送钉钉机器人消息提醒的 Windows 服务程序。

## 功能特性

- 开机自动发送钉钉消息提醒
- 自动等待网络就绪
- 指数退避重试机制
- 优雅的服务启停
- 结构化日志记录（文件 + 控制台）
- 配置灵活支持（Viper）
- 环境变量覆盖配置
- 完整的配置验证
- 基于接口设计，便于测试
- 支持多种钉钉消息类型

## 快速开始

### 1. 配置钉钉机器人

在钉钉群中添加自定义机器人，获取 Webhook 和加签秘钥。

### 2. 编辑配置文件

程序支持 JSON 和 TOML 两种配置文件格式，优先加载 `config.toml`。

**JSON 格式 (config.json)：**

```json
{
  "dingtalk": {
    "webhook": "https://oapi.dingtalk.com/robot/send?access_token=你的Token",
    "secret": "SEC你的加签秘钥"
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
}
```

**TOML 格式 (config.toml)：**

```toml
[dingtalk]
webhook = "https://oapi.dingtalk.com/robot/send?access_token=你的Token"
secret = "SEC你的加签秘钥"

[retry]
max_attempts = 5
init_delay = "2s"
max_delay = "30s"

[network]
wait_timeout = "120s"

[log]
filename = "service.log"
max_size = 10
max_backups = 3
max_age = 30
compress = true
level = "info"
console = true
```

### 3. 编译程序

```bash
# 使用 Makefile 构建
make build

# 或直接编译
go build -o DingtalkBootNotify.exe
```

### 4. 安装服务

以管理员身份运行 cmd：

```cmd
DingtalkBootNotify.exe install
DingtalkBootNotify.exe start
```

### 5. 服务控制

```cmd
# 停止服务
DingtalkBootNotify.exe stop

# 卸载服务
DingtalkBootNotify.exe uninstall
```

## 配置说明

### 配置文件

程序使用 [Viper](https://github.com/spf13/viper) 库管理配置，支持 JSON 和 TOML 两种格式。

配置文件加载顺序：
1. 优先查找 `config.toml`
2. 如果不存在，查找 `config.json`

支持的文件格式：
- `.json`: JSON 格式配置文件
- `.toml`: TOML 格式配置文件

### DingTalk 配置

| 字段 | 说明 | 必需 |
|------|------|------|
| webhook | 钉钉机器人 Webhook URL | 是 |
| secret | 钉钉机器人加签秘钥 | 否 |

### 重试配置

| 字段 | 说明 | 默认值 |
|------|------|--------|
| max_attempts | 最大重试次数 | 5 |
| init_delay | 初始重试延迟 | 2s |
| max_delay | 最大重试延迟 | 30s |

### 网络配置

| 字段 | 说明 | 默认值 |
|------|------|--------|
| wait_timeout | 等待网络超时时间 | 120s |

### 日志配置

| 字段 | 说明 | 默认值 |
|------|------|--------|
| filename | 日志文件名 | service.log |
| max_size | 单个日志文件最大大小(MB) | 10 |
| max_backups | 保留的旧日志文件数量 | 3 |
| max_age | 旧日志文件保留天数 | 30 |
| compress | 是否压缩旧日志 | true |
| level | 日志级别 | info |
| console | 是否输出到控制台 | false |

### 环境变量

支持通过环境变量覆盖配置（需要 `DINGTALK_` 前缀）：

| 环境变量 | 对应配置 |
|----------|----------|
| `DINGTALK_WEBHOOK` | dingtalk.webhook |
| `DINGTALK_SECRET` | dingtalk.secret |
| `DINGTALK_RETRY_MAX_ATTEMPTS` | retry.max_attempts |
| `DINGTALK_RETRY_INIT_DELAY` | retry.init_delay |
| `DINGTALK_RETRY_MAX_DELAY` | retry.max_delay |
| `DINGTALK_NETWORK_WAIT_TIMEOUT` | network.wait_timeout |
| `DINGTALK_LOG_CONSOLE` | log.console |

## 支持的钉钉消息类型

### Text 消息
- 支持文本内容
- 支持 @ 特定用户和 @ 所有人

### Markdown 消息
- 支持标题、引用、加粗、斜体
- 支持链接、图片
- 支持无序列表、有序列表

### Link 消息
- 支持标题、内容、链接
- 支持可选的图片

### ActionCard 消息
- 支持标题、内容（支持 Markdown）
- 支持单按钮模式
- 支持按钮排列方向

### FeedCard 消息
- 支持多个链接项（最多 10 个）
- 每个链接项包含标题、链接、图片

## 项目结构

```
dingtalk-boot-notify/
├── main.go                  # 主程序入口
├── go.mod                   # Go 模块定义
├── go.sum                   # 依赖版本锁定
├── Makefile                 # 构建脚本
├── config.json              # 配置文件 (JSON 格式)
├── config.toml              # 配置文件 (TOML 格式)
├── config.example.json      # JSON 配置示例
├── config.example.toml      # TOML 配置示例
├── README.md                # 项目文档
└── internal/
    ├── config/              # 配置管理 (Viper)
    │   ├── config.go        # 配置加载、验证
    │   └── config_test.go   # 配置测试
    ├── dingtalk/           # 钉钉客户端
    │   ├── dingtalk.go      # HTTP Client 实现
    │   └── dingtalk_test.go # 钉钉测试
    ├── logger/             # 日志管理
    │   ├── logger.go       # Zap + Lumberjack
    │   └── logger_test.go  # 日志测试
    └── worker/             # 业务逻辑
        ├── worker.go       # 网络检测、消息发送
        └── worker_test.go  # 工作器测试
```

## 架构设计

### 接口设计

项目采用接口设计，便于单元测试：

```go
// dingtalk.Sender 接口
type Sender interface {
    Send(ctx context.Context, message string) error
    SendMarkdown(ctx context.Context, title, text string, atInfo *AtInfo) error
    SendLink(ctx context.Context, title, text, messageURL, picURL string) error
    SendActionCard(ctx context.Context, title, text, singleTitle, singleURL, btnOrientation string) error
    SendFeedCard(ctx context.Context, links []FeedCardLink) error
}

// dingtalk.HTTPClient 接口
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}
```

### 核心流程

1. **服务启动** → 加载配置 → 初始化日志
2. **等待网络** → 使用 ticker 定期检测网络连通性
3. **发送消息** → 指数退避重试机制
4. **优雅关闭** → 响应系统信号 (SIGINT/SIGTERM)

### 配置验证

启动时自动验证配置合法性：

- Webhook URL 格式验证
- 重试参数合法性检查
- 日志配置完整性检查

## 构建脚本

使用 `make` 命令管理构建和服务：

| 命令 | 说明 |
|------|------|
| `make build` | 构建可执行文件 |
| `make build-debug` | 构建调试版本 |
| `make test` | 运行测试 |
| `make vet` | 运行代码检查 |
| `make tidy` | 清理并更新依赖 |
| `make clean` | 清理构建产物 |
| `make run` | 运行程序（用于测试） |
| `make install` | 安装服务 |
| `make uninstall` | 卸载服务 |
| `make start` | 启动服务 |
| `make stop` | 停止服务 |
| `make restart` | 重启服务 |
| `make status` | 查看服务状态 |
| `make help` | 查看帮助 |

## 测试覆盖

| 包 | 测试数 | 状态 |
|----|--------|------|
| config | 15 | ✅ 全部通过 |
| dingtalk | 17 | ✅ 全部通过 |
| logger | 5 | ✅ 全部通过 |
| worker | 11 | ✅ 全部通过 |
| **总计** | **48** | **✅ 全部通过** |

## 技术栈

- Go 1.21+
- Viper - 配置管理
- zap - 结构化日志
- lumberjack - 日志轮转
- kardianos/service - Windows 服务框架

## 许可证

MIT License