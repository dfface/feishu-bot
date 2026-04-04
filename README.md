# Feishu Bot 飞书机器人平台

一个高度可扩展的飞书机器人平台，基于 Go 语言开发，支持长连接接入，轻松接入多个机器人完成不同任务。

## 特性

- 🚀 **模块化架构**：清晰的代码结构，易于扩展和维护
- 🔌 **插件式机器人**：通过接口实现，轻松添加新的机器人
- 📱 **飞书长连接**：支持 WebSocket 长连接模式，无需公网 IP
- 📝 **Memos 集成**：内置 Memos 机器人，支持文本、图片、视频等附件
- ⚙️ **灵活配置**：支持 YAML 配置文件和环境变量
- 📊 **结构化日志**：使用 Zap 日志库，支持 JSON 和 Console 格式

## 项目结构

```
feishu-bot/
├── cmd/
│   └── feishu-bot/          # 主程序入口
│       └── main.go
├── internal/
│   └── bots/                 # 机器人实现
│       └── memos_bot.go      # Memos 机器人
├── pkg/
│   ├── bot/                  # 机器人基础框架
│   │   └── bot.go
│   ├── config/               # 配置管理
│   │   └── config.go
│   └── memos/                # Memos 客户端
│       └── client.go
├── config.yaml.example       # 配置文件示例
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## 快速开始

### 前置要求

- Go 1.21+
- 飞书企业自建应用
- Memos 服务

### 1. 克隆项目

```bash
git clone <repository-url>
cd feishu-bot
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 配置

复制配置文件示例：

```bash
cp config.yaml.example config.yaml
```

编辑 `config.yaml` 文件，填入你的配置信息：

```yaml
feishu:
  app_id: "cli_xxxxxxxxxx"           # 飞书应用 ID
  app_secret: "xxxxxxxxxx"           # 飞书应用密钥
  verification_token: ""              # 验证令牌（可选）
  encrypt_key: ""                     # 加密密钥（可选）
  use_websocket: true                 # 使用 WebSocket 长连接

memos:
  base_url: "http://localhost:5230"  # Memos 服务地址
  access_token: "your_token_here"     # Memos 访问令牌
  default_visibility: "PRIVATE"        # 默认可见性

server:
  port: 8080                          # 监听端口

log:
  level: "info"                        # 日志级别
  format: "json"                       # 日志格式
```

### 4. 获取飞书应用凭证

1. 访问 [飞书开放平台](https://open.feishu.cn/)
2. 创建企业自建应用
3. 在「凭证与基础信息」中获取 App ID 和 App Secret
4. 在「事件订阅」中配置：
   - 请求地址：`https://your-domain.com/webhook/event`
   - 订阅事件：接收消息 `im.message.receive_v1`

### 5. 获取 Memos 访问令牌

1. 登录你的 Memos 服务
2. 在设置中创建访问令牌
3. 复制令牌并填入配置文件

### 6. 运行

```bash
go run cmd/feishu-bot/main.go
```

或者使用配置文件：

```bash
go run cmd/feishu-bot/main.go --config config.yaml
```

## 配置说明

### 环境变量

也可以使用环境变量配置，优先级高于配置文件：

```bash
export FEISHU_BOT_FEISHU_APP_ID="cli_xxxxxxxxxx"
export FEISHU_BOT_FEISHU_APP_SECRET="xxxxxxxxxx"
export FEISHU_BOT_MEMOS_BASE_URL="http://localhost:5230"
export FEISHU_BOT_MEMOS_ACCESS_TOKEN="your_token"
```

### 配置项详解

| 配置项 | 说明 | 必填 |
|--------|------|------|
| feishu.app_id | 飞书应用 ID | 是 |
| feishu.app_secret | 飞书应用密钥 | 是 |
| feishu.verification_token | 验证令牌 | 否 |
| feishu.encrypt_key | 加密密钥 | 否 |
| feishu.use_websocket | 使用 WebSocket | 否 |
| memos.base_url | Memos 服务地址 | 是 |
| memos.access_token | Memos 访问令牌 | 是 |
| memos.default_visibility | 默认可见性 | 否 |
| server.port | 监听端口 | 否 |
| log.level | 日志级别 | 否 |
| log.format | 日志格式 | 否 |

## Memos 机器人使用

### 功能

- ✅ 文本消息：直接保存为 Memo
- ✅ 图片：上传到 Memos 并嵌入
- ✅ 文件：上传到 Memos 并链接
- ✅ 音频：上传到 Memos
- ✅ 视频：上传到 Memos

### 使用方法

1. 在飞书中添加机器人为好友
2. 直接发送消息给机器人
3. 消息会自动保存到你的 Memos

## 架构设计

### 核心接口

```go
// Bot 机器人接口
type Bot interface {
    Name() string
    HandleMessage(ctx context.Context, event *P2MessageReceiveV1) error
    HandleCardAction(ctx context.Context, event *CardActionEvent) error
}
```

### 扩展新机器人

1. 在 `internal/bots/` 下创建新的机器人实现
2. 实现 `Bot` 接口
3. 在 `main.go` 中注册机器人

示例：

```go
// internal/bots/my_bot.go
package bots

import (
    "context"
    "github.com/dfface/feishu-bot/pkg/bot"
)

type MyBot struct {
    name string
}

func NewMyBot() *MyBot {
    return &MyBot{name: "mybot"}
}

func (b *MyBot) Name() string {
    return b.name
}

func (b *MyBot) HandleMessage(ctx context.Context, event *bot.P2MessageReceiveV1) error {
    // 处理消息
    return nil
}

func (b *MyBot) HandleCardAction(ctx context.Context, event *bot.CardActionEvent) error {
    // 处理卡片交互
    return nil
}
```

注册机器人：

```go
// cmd/feishu-bot/main.go
myBot := bots.NewMyBot()
botManager.RegisterBot(myBot)
```

## 开发指南

### 本地开发

1. 复制配置文件
2. 填入开发环境配置
3. 运行 `go run cmd/feishu-bot/main.go`

### 代码规范

- 遵循 Go 官方代码规范
- 使用有意义的变量和函数名
- 添加必要的注释
- 编写测试用例

## 部署

### 使用 Docker

创建 `Dockerfile`：

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o feishu-bot ./cmd/feishu-bot

FROM alpine:latest
COPY --from=builder /app/feishu-bot /usr/local/bin/
ENTRYPOINT ["feishu-bot"]
```

构建并运行：

```bash
docker build -t feishu-bot .
docker run -d --name feishu-bot \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -p 8080:8080 \
  feishu-bot --config /app/config.yaml
```

### 使用 Docker Compose

创建 `docker-compose.yml`：

```yaml
version: '3'
services:
  feishu-bot:
    build: .
    volumes:
      - ./config.yaml:/app/config.yaml
    ports:
      - "8080:8080"
    restart: unless-stopped
```

## 常见问题

### Q: 如何获取飞书的事件推送？

A: 有两种方式：
1. WebSocket 长连接（推荐）：无需公网 IP，设置 `use_websocket: true`
2. Webhook：需要公网 IP，在飞书开放平台配置回调地址

### Q: Memos 支持哪些可见性？

A: 支持三种可见性：
- `PRIVATE`：私有（默认）
- `PROTECTED`：保护
- `PUBLIC`：公开

### Q: 如何添加更多机器人？

A: 参考「扩展新机器人」章节，实现 `Bot` 接口并注册即可。

## 技术栈

- **语言**：Go 1.21+
- **飞书 SDK**：[larksuite/oapi-sdk-go](https://github.com/larksuite/oapi-sdk-go)
- **配置管理**：[spf13/viper](https://github.com/spf13/viper)
- **日志**：[uber-go/zap](https://github.com/uber-go/zap)

## 参考文档

- [飞书开放平台文档](https://open.feishu.cn/document)
- [飞书事件订阅指南](https://open.feishu.cn/document/event-subscription-guide/callback-subscription/callback-overview)
- [Memos 官方文档](https://www.usememos.com/)
- [telegram-integration](https://github.com/usememos/telegram-integration)

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
