package features

import (
	"context"
	"strings"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"

	"github.com/dfface/feishu-bot/pkg/bot"
	"github.com/dfface/feishu-bot/pkg/config"
	"github.com/dfface/feishu-bot/pkg/message"
)

// EchoFeature 回声功能
type EchoFeature struct {
	id          string
	name        string
	description string
	prefix      string
	baseBot     *bot.BaseBot
	logger      *zap.Logger
}

// NewEchoFeature 创建回声功能
func NewEchoFeature() *EchoFeature {
	return &EchoFeature{
		id:          "echo",
		name:        "回声功能",
		description: "原样回复消息",
		prefix:      "!echo",
	}
}

// ID 返回功能ID
func (f *EchoFeature) ID() string {
	return f.id
}

// Name 返回功能名称
func (f *EchoFeature) Name() string {
	return f.name
}

// Description 返回功能描述
func (f *EchoFeature) Description() string {
	return f.description
}

// MatchPrefix 返回匹配前缀
func (f *EchoFeature) MatchPrefix() string {
	return f.prefix
}

// Initialize 初始化功能
func (f *EchoFeature) Initialize(featureConfig *config.FeatureConfig) error {
	if prefix, ok := featureConfig.Config["prefix"].(string); ok {
		f.prefix = prefix
	}
	return nil
}

// SetBaseBot 设置基础机器人
func (f *EchoFeature) SetBaseBot(baseBot *bot.BaseBot) {
	f.baseBot = baseBot
	f.logger = baseBot.Logger
}

// HandleMessage 处理消息
func (f *EchoFeature) HandleMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	msg := event.Event.Message
	sender := event.Event.Sender

	f.logger.Info("Processing echo message",
		zap.String("message_id", *msg.MessageId),
		zap.String("sender_id", *sender.SenderId.OpenId),
	)

	// 处理消息
	msgContent, err := f.baseBot.MsgProcessor.Process(ctx, msg)
	if err != nil {
		f.logger.Error("Failed to process message", zap.Error(err))
		return f.baseBot.SendText(ctx, *sender.SenderId.OpenId, "消息处理失败")
	}

	// 处理命令前缀
	text := msgContent.Text
	if strings.HasPrefix(text, f.prefix) {
		text = strings.TrimPrefix(text, f.prefix)
		text = strings.TrimSpace(text)
	}

	// 根据消息类型回复
	switch msgContent.Type {
	case message.MessageTypePost:
		// 富文本消息
		return f.replyRichText(ctx, *msg.MessageId, msgContent)
	default:
		// 其他消息类型
		replyContent := text + "\n已收到"
		return f.baseBot.SendText(ctx, *sender.SenderId.OpenId, replyContent)
	}
}

// replyRichText 回复富文本消息
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
		builder.NewLine()
	}

	// 添加"已收到"
	builder.AddText("\n已收到")

	// 回复消息
	return f.baseBot.ReplyRichText(ctx, messageID, builder)
}
