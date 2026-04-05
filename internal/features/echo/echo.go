// features Echo 功能实现包
//
// 此包实现了 Echo 回声功能，用于原样回复用户发送的消息。
// 这是一个简单的测试功能，用于验证机器人是否正常工作。
// Echo 功能支持文本消息和富文本消息的回复。
package features

import (
	"context"

	"go.uber.org/zap"

	"github.com/dfface/feishu-bot/internal/bot"
	"github.com/dfface/feishu-bot/internal/config"
	"github.com/dfface/feishu-bot/internal/logger"
	"github.com/dfface/feishu-bot/internal/message"
)

// EchoFeature 回声功能
// 实现 Feature 接口，提供原样回复消息的能力
//
// 此功能支持：
// - 文本消息回复
// - 富文本消息回复
// - 自定义命令前缀
type EchoFeature struct {
	id          string       // 功能唯一标识符
	name        string       // 功能名称
	description string       // 功能描述
	prefix      string       // 匹配前缀，用于识别命令
	baseBot     *bot.BaseBot // 基础机器人实例，提供消息处理和发送能力
}

// NewEchoFeature 创建回声功能
//
// 此函数创建一个新的 Echo 功能实例，设置默认的 ID、名称、描述和前缀。
// 实际的配置初始化在 Initialize 方法中完成。
//
// 返回值：
// - *EchoFeature：创建的 Echo 功能实例
func NewEchoFeature() *EchoFeature {
	return &EchoFeature{
		id:          "echo",
		name:        "回声功能",
		description: "原样回复消息",
		prefix:      "!echo",
	}
}

// ID 返回功能ID
//
// 返回值：
// - string：功能的唯一标识符 "echo"
func (f *EchoFeature) ID() string {
	return f.id
}

// Name 返回功能名称
//
// 返回值：
// - string：功能的名称 "回声功能"
func (f *EchoFeature) Name() string {
	return f.name
}

// Description 返回功能描述
//
// 返回值：
// - string：功能的描述 "原样回复消息"
func (f *EchoFeature) Description() string {
	return f.description
}

// MatchPrefix 返回匹配前缀
//
// 匹配前缀用于识别消息是否应该由此功能处理。
// 当消息文本以 "!echo" 开头时，该功能将被调用。
//
// 返回值：
// - string：匹配前缀 "!echo"
func (f *EchoFeature) MatchPrefix() string {
	return f.prefix
}

// Initialize 初始化功能
//
// 此方法在功能注册时被调用，用于初始化功能所需的配置和资源。
// 主要完成以下工作：
// 1. 从配置中读取自定义前缀（如果有）
//
// 参数：
// - featureConfig：功能配置，包含功能所需的配置信息
//
// 返回值：
// - error：初始化过程中的错误，成功则返回 nil
func (f *EchoFeature) Initialize(featureConfig *config.FeatureConfig) error {
	// 覆盖写死的 Name、Description
	if featureConfig.Name != "" {
		f.name = featureConfig.Name
	}
	if featureConfig.Description != "" {
		f.description = featureConfig.Description
	}

	if prefix, ok := featureConfig.Config["prefix"].(string); ok {
		f.prefix = prefix
	}
	return nil
}

// SetBaseBot 设置基础机器人
//
// 此方法是 Feature 接口的必需方法，用于为功能提供访问机器人基础功能的能力。
// 功能通过 BaseBot 可以访问消息处理器、消息发送器、文件上传器等组件。
//
// 参数：
// - baseBot：基础机器人实例
func (f *EchoFeature) SetBaseBot(baseBot *bot.BaseBot) {
	f.baseBot = baseBot
}

// HandleMessage 处理消息
//
// 此方法是功能的核心，负责处理接收到的消息。
// 主要完成以下工作：
// 1. 根据消息类型回复（文本消息或富文本消息）
//
// 参数：
// - ctx：上下文，用于控制请求的生命周期
// - msgContent：消息内容，包含解析后的消息文本、消息 ID 和发送者信息等
//
// 返回值：
// - error：处理过程中的错误，成功则返回 nil
func (f *EchoFeature) HandleMessage(ctx context.Context, msgContent *message.MessageContent) error {
	logger.Info("Processing echo message",
		zap.String("message_id", msgContent.ID),
		zap.String("sender_id", msgContent.SenderID),
	)

	// 根据消息类型回复
	switch msgContent.Type {
	case message.MessageTypePost:
		// 富文本消息
		return f.replyRichText(ctx, msgContent.ID, msgContent)
	default:
		// 其他消息类型
		replyContent := msgContent.Text + "\n已收到"
		return f.baseBot.SendText(ctx, msgContent.SenderID, replyContent)
	}
}

// replyRichText 回复富文本消息
//
// 此方法用于回复富文本消息，保持原有的格式和样式。
// 主要完成以下工作：
// 1. 创建富文本构建器
// 2. 设置标题（如果有）
// 3. 复制原始消息的所有内容（文本、链接、@、图片、媒体、表情、分割线、代码块、Markdown）
// 4. 添加"已收到"标记
// 5. 回复消息
//
// 参数：
// - ctx：上下文，用于控制请求的生命周期
// - messageID：要回复的消息 ID
// - msgContent：消息内容，包含富文本信息
//
// 返回值：
// - error：回复过程中的错误，成功则返回 nil
func (f *EchoFeature) replyRichText(ctx context.Context, messageID string, msgContent *message.MessageContent) error {
	if msgContent.RichText == nil {
		// 这里应该从 event 中获取 sender_id，暂时使用一个空字符串
		return f.baseBot.SendText(ctx, "", msgContent.Text+"\n已收到")
	}

	// 创建富文本构建器
	builder := message.NewRichTextMessageBuilder()

	// 设置标题
	if msgContent.RichText.Title != "" {
		builder.SetTitle(msgContent.RichText.Title)
	}

	// 复制内容
	for _, line := range msgContent.RichText.Content {
		for _, elem := range line {
			switch elem.Tag {
			case string(message.RichTextTagText):
				builder.AddTextWithStyle(elem.Text, elem.Style...)
			case string(message.RichTextTagA):
				builder.AddLinkWithStyle(elem.Text, elem.Href, elem.Style...)
			case string(message.RichTextTagAt):
				builder.AddAtWithStyle(elem.UserId, elem.UserName, elem.Style...)
			case string(message.RichTextTagImg):
				builder.AddImage(elem.ImageKey)
			case string(message.RichTextTagMedia):
				builder.AddMedia(elem.FileKey, elem.ImageKey)
			case string(message.RichTextTagEmotion):
				builder.AddEmotion(message.EmojiType(elem.EmojiType))
			case string(message.RichTextTagHr):
				builder.AddHr()
			case string(message.RichTextTagCodeBlock):
				builder.AddCodeBlock(elem.Text, elem.Language)
			case string(message.RichTextTagMd):
				builder.AddMd(elem.Text)
			}
		}
		builder.NewParagraph()
	}

	// 添加"已收到"
	builder.AddText("\n已收到")

	// 回复消息
	return f.baseBot.ReplyRichText(ctx, messageID, builder)
}
