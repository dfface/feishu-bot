# Feishu Bot 飞书机器人平台

## 📋 项目简介

Feishu Bot 是一个高度可扩展的飞书机器人平台，基于 Go 语言开发，通过配置文件轻松管理多个机器人和功能，可根据发送给机器人的消息，执行不同的功能操作，例如将消息写入 [Memos](https://usememos.com/) 系统中。

- **配置驱动**：通过 YAML 配置文件定义机器人和功能的映射关系
- **模块化架构**：清晰的代码结构，易于扩展和维护
- **功能丰富**：内置 Echo、Memos 等功能，支持文本、图片、文件等多种消息类型

## ✨ 核心特性

```
┌─────────────────────────────────────────┐
│             核心概念架构图                │
├─────────────────────────────────────────┤
│                                         │
│           ┌───────────────┐             │
│           │功能 (Features) │             │
│           ├───────────────┤             │
│           │  ┌─────────┐  │             │
│           │  │  Echo   │  │             │
│           │  └─────────┘  │             │
│           │  ┌─────────┐  │             │
│           │  │ Memos   │  │             │
│           │  └─────────┘  │             │
│           │  ┌─────────┐  │             │
│           │  │  ...    │  │             │
│           │  └─────────┘  │             │
│           └───────────────┘             │
│                 ▲                       │
│                 │ 映射关系               │
│                 ▼                       │
│           ┌───────────────┐             │
│           │ 机器人 (Bots)  │             │
│           ├───────────────┤             │
│           │  ┌─────────┐  │             │
│           │  │  机器人1 │  │             │
│           │  └─────────┘  │             │
│           │  ┌─────────┐  │             │
│           │  │  机器人2 │  │             │
│           │  └─────────┘  │             │
│           │  ┌─────────┐  │             │
│           │  │  ...    │  │             │
│           │  └─────────┘  │             │
│           └───────────────┘             │
│                                         │
└─────────────────────────────────────────┘

说明：
- 一个机器人可以映射多个功能
- 一个功能可以被多个机器人使用
- 通过配置文件定义映射关系
```

- 📋 **当前支持的功能**：
  - 🔄 **Echo 测试**：内置回声功能，向用户发送相同的消息，用于测试和调试
  - 📝 **Memos 集成**：支持将消息保存到 Memos，支持富文本、markdown、图片、文件等

## 🚀 快速开始

### 安装方式

#### 方式 1：从 Release 下载（推荐）

1. 前往 GitHub 仓库的 [Releases](https://github.com/dfface/feishu-bot/releases) 页面
2. 下载对应平台的可执行文件
3. 解压并运行

#### 方式 2：使用 Docker

**构建镜像**：

```bash
docker pull dfface/feishu-bot:latest
```

**运行容器**：

```bash
docker run -d --name feishu-bot \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  --restart unless-stopped \
  dfface/feishu-bot:latest
```

**使用 Docker Compose**：

```bash
docker compose up -d
```

#### 方式 3：手动编译

**前置要求**：
- Go 1.25+

1. **克隆项目**

```bash
git clone <repository-url>
cd feishu-bot
```

2. **安装依赖**

```bash
go mod tidy
```

3. **编译**

```bash
go build -o feishu-bot ./cmd/feishu-bot
```

### 配置

#### 前置要求

- 飞书企业自建应用：
  - 应用类型：机器人
  - 应用权限：消息接收、用户进入与机器人的会话
  - 事件订阅： `im.message.receive_v1`、`im.chat.access_event.bot_p2p_chat_entered_v1`
- Memos 服务（仅用于 Memos 功能）

无论使用哪种安装方式，都需要配置 `config.yaml` 文件：

#### 编写配置

复制配置文件示例：

```bash
cp config.yaml.example config.yaml
```

编辑 `config.yaml` 文件，填入你的配置信息：

```yaml
# 日志配置
log:
  level: "info"
  format: "json"

# 功能配置
features:
  - id: "echo"
    name: "回声功能"
    enabled: true
    config:
      prefix: "!echo"

  - id: "memos"
    name: "Memos 功能"
    enabled: true
    config:
      prefix: "!memos"
      memos:
        base_url: "http://localhost:5230"
        access_token: "your_memos_token"
        default_visibility: "PRIVATE"

# 机器人配置
bots:
  - id: "multi-bot"
    name: "多功能机器人"
    enabled: true
    welcome_message_enabled: true  # 是否启用欢迎消息
    feishu:
      app_id: "cli_xxxxxxxxxx"
      app_secret: "xxxxxxxxxx"
      use_websocket: true
    features:
      - feature_id: "echo"
        default: true  # 设置为默认功能
      - feature_id: "memos"
```

### 运行

**使用可执行文件**：

```bash
./feishu-bot --config config.yaml
```

**使用 Go 运行**（开发模式）：

```bash
go run cmd/feishu-bot/main.go --config config.yaml
```

## ⚙️ 配置指南

### 配置文件结构

配置文件采用 YAML 格式，主要包含以下部分：

1. **日志配置**：日志级别和格式设置
2. **功能配置**：定义可用的功能模块
3. **机器人配置**：定义机器人及其功能映射

### 功能配置

| 配置项 | 说明 | 示例 |
|--------|------|------|
| id | 功能唯一标识符 | `echo` |
| name | 功能名称 | `回声功能` |
| enabled | 是否启用 | `true` |
| config | 功能配置（map 格式） | 见下方示例 |

**功能配置示例**：

```yaml
# Echo 功能配置
config:
  prefix: "!echo"  # 命令前缀

# Memos 功能配置
config:
  prefix: "!memos"  # 命令前缀
  memos:           # Memos 特定配置
    base_url: "http://localhost:5230"
    access_token: "your_memos_token"
    default_visibility: "PRIVATE"
```

### 机器人配置

| 配置项 | 说明 | 示例 |
|--------|------|------|
| id | 机器人唯一标识符 | `multi-bot` |
| name | 机器人名称 | `多功能机器人` |
| enabled | 是否启用 | `true` |
| welcome_message_enabled | 是否启用欢迎消息 | `true` |
| feishu | 飞书应用配置 | 见下方示例 |
| features | 功能映射列表 | 见下方示例 |

**飞书配置示例**：

```yaml
feishu:
  app_id: "cli_xxxxxxxxxx"
  app_secret: "xxxxxxxxxx"
  verification_token: ""
  encrypt_key: ""
  use_websocket: true
```

**功能映射示例**：

```yaml
features:
  - feature_id: "echo"
    default: true  # 设置为默认功能
  - feature_id: "memos"
```

## 📱 使用方法

### 多功能机器人

**功能说明**：
- `!echo`：回声功能，原样回复消息
- `!memos`：Memos 功能，保存消息到 Memos

**使用示例**：
1. 发送 `!echo 你好` → 触发回声功能，回复 "你好\n已收到"
2. 发送 `!memos 测试记录` → 触发 Memos 功能，保存到 Memos 并回复 "已保存到 Memos"
3. 发送任意内容 → 触发默认功能（echo）

### 功能详情

#### Echo 功能

- ✅ 文本消息：原样回复并添加「已收到」
- ✅ 富文本消息：原样回复并添加「已收到」
- ✅ 图片：重新上传并回复
- ✅ 表情反应：自动添加相同的表情反应

#### Memos 功能

- ✅ 文本消息：直接保存为 Memo
- ✅ 富文本消息：转换为 Markdown 格式保存
- ✅ 图片：下载并上传到 Memos
- ✅ 文件：下载并上传到 Memos
- ✅ 音频：下载并上传到 Memos
- ✅ 视频：下载并上传到 Memos

## 🏗️ 架构设计

### 核心概念

- **功能（Feature）**：独立的功能模块，如 echo、memos 等
- **机器人（Bot）**：可以包含多个功能的机器人实例
- **配置驱动**：通过配置文件定义功能和机器人的映射关系

### 智能功能匹配

- **前缀匹配**：根据消息前缀匹配相应的功能（如 `!echo`、`!memos`）
- **默认功能**：如果没有匹配的前缀，使用默认功能
- **单功能机器人**：只有一个功能时，直接使用该功能

### 项目结构

```
feishu-bot/
├── cmd/                  # 主程序入口
│   └── feishu-bot/
│       └── main.go
├── internal/             # 内部实现（不可被外部导入）
│   ├── bot/              # 机器人基础框架
│   ├── bots/             # 机器人实现
│   │   ├── bot.go        # 统一的 Bot 实现
│   │   └── bot_factory.go # 机器人工厂
│   ├── config/           # 配置管理
│   ├── converter/        # 消息转换器
│   ├── features/         # 功能模块
│   │   ├── echo/         # 回声功能
│   │   ├── memos/        # Memos 功能
│   │   └── features.go   # 功能接口和注册中心
│   └── message/          # 消息处理核心
├── third_party/          # 第三方服务客户端
│   └── memos/            # Memos 客户端
├── config.yaml.example    # 配置文件示例
└── README.md             # 项目说明
```

## 🛠️ 开发指南

### 扩展新功能

1. **创建功能实现**：在 `internal/features/` 下创建新的功能目录
2. **实现 Feature 接口**：实现 `ID()`、`Name()`、`Description()`、`MatchPrefix()`、`Initialize()`、`SetBaseBot()` 和 `HandleMessage()` 方法
3. **注册功能**：在 `internal/features/features.go` 中注册新功能
4. **配置功能**：在配置文件中添加功能配置

**示例**：创建一个问候功能

```go
// internal/features/hello/hello.go
package hello

import (
    "context"
    larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
    "go.uber.org/zap"
    "github.com/dfface/feishu-bot/internal/bot"
    "github.com/dfface/feishu-bot/internal/config"
)

type HelloFeature struct {
    id          string
    name        string
    description string
    prefix      string
    baseBot     *bot.BaseBot
    logger      *zap.Logger
}

func NewHelloFeature() *HelloFeature {
    return &HelloFeature{
        id:          "hello",
        name:        "问候功能",
        description: "向用户问好",
        prefix:      "!hello",
    }
}

func (f *HelloFeature) ID() string { return f.id }
func (f *HelloFeature) Name() string { return f.name }
func (f *HelloFeature) Description() string { return f.description }
func (f *HelloFeature) MatchPrefix() string { return f.prefix }
func (f *HelloFeature) SetBaseBot(baseBot *bot.BaseBot) {
    f.baseBot = baseBot
    f.logger = baseBot.Logger
}

func (f *HelloFeature) Initialize(featureConfig *config.FeatureConfig) error {
    if prefix, ok := featureConfig.Config["prefix"].(string); ok {
        f.prefix = prefix
    }
    return nil
}

func (f *HelloFeature) SetBaseBot(baseBot *bot.BaseBot) {
    f.baseBot = baseBot
    f.logger = baseBot.Logger
}

func (f *HelloFeature) HandleMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
    sender := event.Event.Sender
    return f.baseBot.SendText(ctx, *sender.SenderId.OpenId, "Hello! 你好！")
}
```

**注册功能**：

```go
// internal/features/features.go
func RegisterFeatures(registry *FeatureRegistry) {
    registry.Register(echo.NewEchoFeature())
    registry.Register(memos.NewMemosFeature())
    registry.Register(hello.NewHelloFeature()) // 添加这一行
}
```

**配置功能**：

```yaml
features:
  - id: "hello_you_defined"
    internal_id: "hello"  # 内部功能 ID，与功能实现中代码写死的 ID 一致
    config:
      prefix: "!hello"
  - id: "echo_you_defined"
    internal_id: "echo"  # 内部功能 ID，与功能实现中代码写死的 ID 一致
    config:
      prefix: "!echo"

bots:
  - id: "multi-bot"
    name: "多功能机器人"
    enabled: true
    features:
      - feature_id: "echo_you_defined"
        default: true
      - feature_id: "hello_you_defined" # 添加这一行
```

### 本地开发

1. 复制配置文件：`cp config.yaml.example config.yaml`
2. 填入开发环境配置
3. 运行：`go run cmd/feishu-bot/main.go --config config.yaml`

## 📦 部署

### 部署建议

- **生产环境**：推荐使用 Docker 部署，便于管理和维护
- **测试环境**：可以使用 Release 下载的可执行文件或手动编译
- **开发环境**：使用 `go run` 直接运行，方便调试

## ❓ 常见问题

### Q: 如何获取飞书的事件推送？

A: 有两种方式：
1. **WebSocket 长连接**（推荐）：无需公网 IP，设置 `use_websocket: true`
2. **Webhook**：需要公网 IP，在飞书开放平台配置回调地址

### Q: Memos 支持哪些可见性？

A: 支持三种可见性：
- `PRIVATE`：私有（默认）
- `PROTECTED`：保护
- `PUBLIC`：公开

### Q: 如何添加更多功能？

A: 参考「扩展新功能」章节，实现 `Feature` 接口并注册功能即可。

### Q: 如何配置多个机器人？

A: 在配置文件的 `bots` 数组中添加多个机器人配置即可。

## 🛠️ 技术栈

- **语言**：Go 1.25+
- **飞书 SDK**：[larksuite/oapi-sdk-go](https://github.com/larksuite/oapi-sdk-go)
- **配置管理**：[spf13/viper](https://github.com/spf13/viper)
- **日志**：[uber-go/zap](https://github.com/uber-go/zap)
- **Memos SDK**：[usememos/memos](https://github.com/usememos/memos)

## 📚 参考文档

- [飞书开放平台文档](https://open.feishu.cn/document)
- [飞书事件订阅指南](https://open.feishu.cn/document/event-subscription-guide/callback-subscription/callback-overview)
- [Memos 官方文档](https://www.usememos.com/)

## 📄 许可证

MIT License

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！
