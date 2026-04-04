package bots

import (
	"context"
	"fmt"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"

	"github.com/dfface/feishu-bot/pkg/bot"
	"github.com/dfface/feishu-bot/pkg/config"
	"github.com/dfface/feishu-bot/pkg/message"
)

// EchoBot 回声机器人 - 收到什么消息就回复什么消息加上"已收到"
type EchoBot struct {
	*bot.BaseBot
	cfg *config.Config
}

// NewEchoBot 创建回声机器人
//
// 参数:
//
//	name - 机器人名称
//	client - 飞书 API 客户端
//	cfg - 应用配置
//	logger - 日志记录器
//
// 返回:
//
//	*EchoBot - 初始化好的回声机器人实例
func NewEchoBot(name string, client *lark.Client, cfg *config.Config, logger *zap.Logger) *EchoBot {
	b := &EchoBot{
		BaseBot: bot.NewBaseBot(name, client, logger),
		cfg:     cfg,
	}

	// 设置事件处理器
	b.OnMessage(b.HandleMessage)
	b.OnReactionCreated(b.HandleReactionCreated)
	b.OnReactionDeleted(b.HandleReactionDeleted)

	return b
}

// HandleMessage 处理消息事件
func (b *EchoBot) HandleMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	msg := event.Event.Message
	sender := event.Event.Sender

	b.Logger.Info("Processing message",
		zap.String("message_id", *msg.MessageId),
		zap.String("sender_id", *sender.SenderId.OpenId),
		zap.String("chat_type", *msg.ChatType),
		zap.String("message_type", *msg.MessageType),
	)

	// 使用消息处理器解析消息
	msgContent, err := b.MsgProcessor.Process(ctx, msg)
	if err != nil {
		b.Logger.Error("Failed to process message", zap.Error(err))
		return b.ReplyText(ctx, *msg.MessageId, fmt.Sprintf("消息处理失败: %v", err))
	}

	// 记录资源信息
	if len(msgContent.Resources) > 0 {
		b.Logger.Info("Message contains resources",
			zap.Int("resource_count", len(msgContent.Resources)),
		)
		for i, resource := range msgContent.Resources {
			b.Logger.Info("Resource info",
				zap.Int("index", i),
				zap.String("type", string(resource.Type)),
				zap.String("file_key", resource.FileKey),
				zap.String("file_name", resource.FileName),
				zap.String("local_path", resource.LocalPath),
				zap.Bool("downloaded", resource.Downloaded),
			)
		}
	}

	// 根据消息类型选择不同的回复方式
	switch msgContent.Type {
	case message.MessageTypePost:
		// 富文本消息，解析并添加"已收到"
		return b.replyRichText(ctx, *msg.MessageId, msgContent)
	default:
		// 其他消息类型，使用文本回复
		replyContent := fmt.Sprintf("%s\n已收到", msgContent.Text)
		return b.ReplyText(ctx, *msg.MessageId, replyContent)
	}
}

// replyRichText 回复富文本消息
func (b *EchoBot) replyRichText(ctx context.Context, messageID string, msgContent *message.MessageContent) error {
	if msgContent.RichText == nil {
		// 如果没有解析后的富文本，回退到文本回复
		replyContent := fmt.Sprintf("%s\n已收到", msgContent.Text)
		return b.ReplyText(ctx, messageID, replyContent)
	}

	// 创建旧 image_key 到新 image_key 的映射
	imageKeyMap := make(map[string]string)

	// 重新上传所有图片资源
	for _, resource := range msgContent.Resources {
		if resource.Type == message.ResourceTypeImage && resource.Downloaded && resource.LocalPath != "" {
			b.Logger.Info("Re-uploading image for rich text",
				zap.String("file_key", resource.FileKey),
				zap.String("local_path", resource.LocalPath))
			newImageKey, err := b.FileUploader.UploadImage(ctx, resource.LocalPath, message.ImageTypeMessage)
			if err != nil {
				b.Logger.Error("Failed to re-upload image",
					zap.String("file_key", resource.FileKey),
					zap.Error(err))
				continue
			}
			imageKeyMap[resource.FileKey] = newImageKey
			b.Logger.Info("Image re-uploaded successfully",
				zap.String("old_key", resource.FileKey),
				zap.String("new_key", newImageKey))
		}
	}

	// 创建富文本构建器
	builder := message.NewRichTextMessageBuilder()

	// 如果有标题，设置标题
	if msgContent.RichText.Title != "" {
		builder.SetTitle(msgContent.RichText.Title)
	}

	// 复制解析后的富文本内容
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
				// 先尝试用映射的 key
				if newImageKey, exists := imageKeyMap[elem.ImageKey]; exists {
					b.Logger.Info("Using mapped image key",
						zap.String("old_key", elem.ImageKey),
						zap.String("new_key", newImageKey))
					builder.AddImage(newImageKey)
				} else {
					// 如果没有映射，直接使用原 image_key
					b.Logger.Info("Using original image key directly",
						zap.String("image_key", elem.ImageKey))
					builder.AddImage(elem.ImageKey)
				}
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
		// 每一行处理完后换行
		builder.NewLine()
	}

	// 添加"已收到"的新行
	builder.AddText("\n已收到")

	// 回复富文本消息
	return b.ReplyRichText(ctx, messageID, builder)
}

// HandleReactionCreated 处理表情反应创建事件
func (b *EchoBot) HandleReactionCreated(ctx context.Context, event *larkim.P2MessageReactionCreatedV1) error {
	reaction := event.Event

	var userID string
	if reaction.UserId != nil && reaction.UserId.OpenId != nil {
		userID = *reaction.UserId.OpenId
	}

	var emojiType string
	if reaction.ReactionType != nil && reaction.ReactionType.EmojiType != nil {
		emojiType = *reaction.ReactionType.EmojiType
	}

	b.Logger.Info("Received reaction created event",
		zap.String("message_id", *reaction.MessageId),
		zap.String("emoji_type", emojiType),
		zap.String("user_id", userID),
	)

	// 同样的表情反应回去
	return b.AddReaction(ctx, *reaction.MessageId, message.EmojiType(emojiType))
}

// HandleReactionDeleted 处理表情反应删除事件
func (b *EchoBot) HandleReactionDeleted(ctx context.Context, event *larkim.P2MessageReactionDeletedV1) error {
	reaction := event.Event

	var userID string
	if reaction.UserId != nil && reaction.UserId.OpenId != nil {
		userID = *reaction.UserId.OpenId
	}

	var emojiType string
	if reaction.ReactionType != nil && reaction.ReactionType.EmojiType != nil {
		emojiType = *reaction.ReactionType.EmojiType
	}

	b.Logger.Info("Received reaction deleted event",
		zap.String("message_id", *reaction.MessageId),
		zap.String("emoji_type", emojiType),
		zap.String("user_id", userID),
	)

	// 这里可以添加删除表情反应的逻辑，但通常不需要

	return nil
}
