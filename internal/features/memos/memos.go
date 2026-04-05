// features Memos 功能实现包
//
// 此包实现了 Memos 功能，用于将飞书消息保存到 Memos 笔记系统。
// Memos 是一个开源的笔记系统，支持快速记录和组织笔记。
// 此功能允许用户通过飞书机器人将消息保存到 Memos，实现跨平台笔记记录。
package features

import (
	"context"
	"os"
	"strings"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"

	"github.com/dfface/feishu-bot/internal/bot"
	"github.com/dfface/feishu-bot/internal/config"
	"github.com/dfface/feishu-bot/internal/converter"
	memos "github.com/dfface/feishu-bot/third_party/memos"
)

// MemosFeature Memos 功能
// 实现 Feature 接口，提供将飞书消息保存到 Memos 的能力
//
// 此功能支持：
// - 文本消息保存
// - 富文本消息转换
// - 图片和文件附件上传
// - 自定义可见性设置
type MemosFeature struct {
	id          string              // 功能唯一标识符
	name        string              // 功能名称
	description string              // 功能描述
	prefix      string              // 匹配前缀，用于识别命令
	baseBot     *bot.BaseBot        // 基础机器人实例，提供消息处理和发送能力
	logger      *zap.Logger         // 日志记录器
	memosClient *memos.Client       // Memos API 客户端
	cfg         *config.Config      // 全局配置
	memosConfig *config.MemosConfig // Memos 特定配置
}

// NewMemosFeature 创建 Memos 功能
//
// 此函数创建一个新的 Memos 功能实例，设置默认的 ID、名称、描述和前缀。
// 实际的配置和客户端初始化在 Initialize 方法中完成。
//
// 返回值：
// - *MemosFeature：创建的 Memos 功能实例
func NewMemosFeature() *MemosFeature {
	return &MemosFeature{
		id:          "memos",
		name:        "Memos 保存",
		description: "保存消息到 Memos",
		prefix:      "!memos",
	}
}

// ID 返回功能ID
//
// 返回值：
// - string：功能的唯一标识符 "memos"
func (f *MemosFeature) ID() string {
	return f.id
}

// Name 返回功能名称
//
// 返回值：
// - string：功能的名称 "Memos 保存"
func (f *MemosFeature) Name() string {
	return f.name
}

// Description 返回功能描述
//
// 返回值：
// - string：功能的描述 "保存消息到 Memos"
func (f *MemosFeature) Description() string {
	return f.description
}

// MatchPrefix 返回匹配前缀
//
// 匹配前缀用于识别消息是否应该由此功能处理。
// 当消息文本以 "!memos" 开头时，该功能将被调用。
//
// 返回值：
// - string：匹配前缀 "!memos"
func (f *MemosFeature) MatchPrefix() string {
	return f.prefix
}

// Initialize 初始化功能
//
// 此方法在功能注册时被调用，用于初始化功能所需的配置和资源。
// 主要完成以下工作：
// 1. 从配置中读取自定义前缀（如果有）
// 2. 从配置中读取 Memos 连接信息（BaseURL、AccessToken、DefaultVisibility）
// 3. 创建 Memos API 客户端
//
// 参数：
// - featureConfig：功能配置，包含功能所需的配置信息
//
// 返回值：
// - error：初始化过程中的错误，成功则返回 nil
func (f *MemosFeature) Initialize(featureConfig *config.FeatureConfig) error {
	if prefix, ok := featureConfig.Config["prefix"].(string); ok {
		f.prefix = prefix
	}

	// 从 Config map 中读取 Memos 配置
	if memosConfigMap, ok := featureConfig.Config["memos"].(map[string]interface{}); ok {
		f.memosConfig = &config.MemosConfig{
			BaseURL:           getStringValue(memosConfigMap, "base_url", ""),
			AccessToken:       getStringValue(memosConfigMap, "access_token", ""),
			DefaultVisibility: getStringValue(memosConfigMap, "default_visibility", "PRIVATE"),
		}

		// 创建 Memos 客户端
		if f.memosConfig.BaseURL != "" && f.memosConfig.AccessToken != "" {
			f.memosClient = memos.NewClient(
				f.memosConfig.BaseURL,
				f.memosConfig.AccessToken,
			)
		}
	}

	return nil
}

// getStringValue 从 map 中获取字符串值
//
// 此辅助函数用于从 map[string]interface{} 中安全地获取字符串值。
// 如果键不存在或值不是字符串类型，返回默认值。
//
// 参数：
// - m：配置 map
// - key：键名
// - defaultValue：默认值
//
// 返回值：
// - string：获取到的字符串值或默认值
func getStringValue(m map[string]interface{}, key, defaultValue string) string {
	if value, ok := m[key].(string); ok {
		return value
	}
	return defaultValue
}

// SetBaseBot 设置基础机器人
//
// 此方法是 Feature 接口的必需方法，用于为功能提供访问机器人基础功能的能力。
// 功能通过 BaseBot 可以访问消息处理器、消息发送器、文件上传器等组件。
//
// 参数：
// - baseBot：基础机器人实例
func (f *MemosFeature) SetBaseBot(baseBot *bot.BaseBot) {
	f.baseBot = baseBot
}

// SetMemosClient 设置 Memos 客户端
//
// 此方法用于设置 Memos API 客户端，主要用于测试和依赖注入。
//
// 参数：
// - client：Memos API 客户端实例
func (f *MemosFeature) SetMemosClient(client *memos.Client) {
	f.memosClient = client
}

// SetConfig 设置配置
//
// 此方法用于设置全局配置，主要用于测试和依赖注入。
//
// 参数：
// - cfg：全局配置实例
func (f *MemosFeature) SetConfig(cfg *config.Config) {
	f.cfg = cfg
}

// SetMemosConfig 设置 Memos 配置
//
// 此方法用于设置 Memos 特定配置，主要用于测试和依赖注入。
//
// 参数：
// - memosConfig：Memos 配置实例
func (f *MemosFeature) SetMemosConfig(memosConfig *config.MemosConfig) {
	f.memosConfig = memosConfig
}

// HandleMessage 处理消息
//
// 此方法是功能的核心，负责处理接收到的消息。
// 主要完成以下工作：
// 1. 解析消息内容
// 2. 移除命令前缀
// 3. 转换消息格式（支持文本、富文本、图片、文件等）
// 4. 上传附件到 Memos
// 5. 创建 Memo
// 6. 清理临时文件
// 7. 回复用户
//
// 参数：
// - ctx：上下文，用于控制请求的生命周期
// - event：消息事件，包含消息内容和发送者信息
//
// 返回值：
// - error：处理过程中的错误，成功则返回 nil
func (f *MemosFeature) HandleMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	msg := event.Event.Message
	sender := event.Event.Sender

	f.logger.Info("Processing memos message",
		zap.String("message_id", *msg.MessageId),
		zap.String("sender_id", *sender.SenderId.OpenId),
	)

	// 处理消息
	msgContent, err := f.baseBot.MsgProcessor.Process(ctx, msg)
	if err != nil {
		f.logger.Error("Failed to process message", zap.Error(err))
		return f.baseBot.ReplyText(ctx, *sender.SenderId.OpenId, "消息处理失败")
	}

	// 处理命令前缀
	text := msgContent.Text
	if strings.HasPrefix(text, f.prefix) {
		text = strings.TrimPrefix(text, f.prefix)
		text = strings.TrimSpace(text)
		msgContent.Text = text
	}

	// 使用转换器转换消息
	content, filePaths, err := converter.NewMemosConverter().ConvertMessageContent(msgContent)
	if err != nil {
		f.logger.Error("Failed to convert message", zap.Error(err))
		return f.baseBot.ReplyText(ctx, *sender.SenderId.OpenId, "消息转换失败")
	}

	// 创建 Memo
	visibility := memos.Visibility(f.memosConfig.DefaultVisibility)
	memo, attachments, err := f.memosClient.CreateMemoWithResources(ctx, content, visibility, filePaths)
	if err != nil {
		f.logger.Error("Failed to create memo", zap.Error(err))
		return f.baseBot.ReplyText(ctx, *sender.SenderId.OpenId, "保存失败")
	}

	// 清理本地文件
	for _, path := range filePaths {
		_ = os.Remove(path)
	}

	f.logger.Info("Memo created",
		zap.String("memo_name", memo.Name),
		zap.Int("attachments", len(attachments)),
	)

	// 回复成功
	return f.baseBot.ReplyText(ctx, *sender.SenderId.OpenId, "已保存到 Memos")
}
