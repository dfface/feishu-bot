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
	msgProcessor message.MessageReceiver
	msgSender    message.MessageSender
	fileUploader message.FileUploader
	cfg          *config.Config
	logger       *zap.Logger
}

// NewEchoBot 创建回声机器人
func NewEchoBot(name string, client *lark.Client, cfg *config.Config, logger *zap.Logger) *EchoBot {
	baseBot := bot.NewBaseBot(name, client, cfg.Feishu.VerificationToken, cfg.Feishu.EncryptKey)

	bot := &EchoBot{
		BaseBot:      baseBot,
		msgProcessor: message.NewProcessor(client, logger),
		msgSender:    message.NewMessageSender(client, logger),
		fileUploader: message.NewFileUploader(client, logger),
		cfg:          cfg,
		logger:       logger,
	}

	// 设置事件处理器
	bot.GetDispatcher().OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
		return bot.HandleMessage(ctx, event)
	})

	// 设置表情反应事件处理器
	bot.GetDispatcher().OnP2MessageReactionCreatedV1(func(ctx context.Context, event *larkim.P2MessageReactionCreatedV1) error {
		return bot.HandleReactionCreated(ctx, event)
	})

	bot.GetDispatcher().OnP2MessageReactionDeletedV1(func(ctx context.Context, event *larkim.P2MessageReactionDeletedV1) error {
		return bot.HandleReactionDeleted(ctx, event)
	})

	return bot
}

// HandleMessage 处理消息事件
func (b *EchoBot) HandleMessage(ctx context.Context, event interface{}) error {
	imEvent := event.(*larkim.P2MessageReceiveV1)
	msg := imEvent.Event.Message
	sender := imEvent.Event.Sender

	b.logger.Info("Processing message",
		zap.String("message_id", *msg.MessageId),
		zap.String("sender_id", *sender.SenderId.OpenId),
		zap.String("chat_type", *msg.ChatType),
		zap.String("message_type", *msg.MessageType),
		zap.Any("full_message", msg),
	)

	// 使用消息处理器解析消息
	msgContent, err := b.msgProcessor.Process(ctx, msg)
	if err != nil {
		b.logger.Error("Failed to process message", zap.Error(err))
		return b.replyText(ctx, *msg.MessageId, fmt.Sprintf("消息处理失败: %v", err))
	}

	// 记录资源信息
	if len(msgContent.Resources) > 0 {
		b.logger.Info("Message contains resources",
			zap.Int("resource_count", len(msgContent.Resources)),
		)
		for i, resource := range msgContent.Resources {
			b.logger.Info("Resource info",
				zap.Int("index", i),
				zap.String("type", resource.Type),
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
		return b.replyText(ctx, *msg.MessageId, replyContent)
	}
}

// HandleCardAction 处理卡片交互事件
func (b *EchoBot) HandleCardAction(ctx context.Context, event interface{}) error {
	// 暂不处理卡片交互
	b.logger.Info("Received card action", zap.Any("event", event))
	return nil
}

// replyRichText 回复富文本消息
func (b *EchoBot) replyRichText(ctx context.Context, messageID string, msgContent *message.MessageContent) error {
	if msgContent.RichText == nil {
		// 如果没有解析后的富文本，回退到文本回复
		replyContent := fmt.Sprintf("%s\n已收到", msgContent.Text)
		return b.replyText(ctx, messageID, replyContent)
	}

	// 创建旧 image_key 到新 image_key 的映射
	imageKeyMap := make(map[string]string)

	// 重新上传所有图片资源
	for _, resource := range msgContent.Resources {
		if resource.Type == "image" && resource.Downloaded && resource.LocalPath != "" {
			b.logger.Info("Re-uploading image for rich text", zap.String("file_key", resource.FileKey), zap.String("local_path", resource.LocalPath))
			newImageKey, err := b.fileUploader.UploadImage(ctx, resource.LocalPath, message.ImageTypeMessage)
			if err != nil {
				b.logger.Error("Failed to re-upload image", zap.String("file_key", resource.FileKey), zap.Error(err))
				continue
			}
			imageKeyMap[resource.FileKey] = newImageKey
			b.logger.Info("Image re-uploaded successfully", zap.String("old_key", resource.FileKey), zap.String("new_key", newImageKey))
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
		lineElements := make([]message.RichTextElement, 0, len(line))
		for _, elem := range line {
			switch elem.Tag {
			case "text":
				lineElements = append(lineElements, &message.RichTextText{Text: elem.Text, UnEscape: false, Style: elem.Style})
			case "a":
				lineElements = append(lineElements, &message.RichTextA{Text: elem.Text, Href: elem.Href, UnEscape: false, Style: elem.Style})
			case "at":
				lineElements = append(lineElements, &message.RichTextAt{UserId: elem.UserId, UserName: elem.UserName, Style: elem.Style})
			case "img":
				// 先尝试用映射的 key
				if newImageKey, exists := imageKeyMap[elem.ImageKey]; exists {
					b.logger.Info("Using mapped image key", zap.String("old_key", elem.ImageKey), zap.String("new_key", newImageKey))
					lineElements = append(lineElements, &message.RichTextImg{ImageKey: newImageKey})
				} else {
					// 如果没有映射，直接使用原 image_key
					b.logger.Info("Using original image key directly", zap.String("image_key", elem.ImageKey))
					lineElements = append(lineElements, &message.RichTextImg{ImageKey: elem.ImageKey})
				}
			case "media":
				lineElements = append(lineElements, &message.RichTextMedia{FileKey: elem.FileKey, ImageKey: elem.ImageKey})
			case "emotion":
				lineElements = append(lineElements, &message.RichTextEmotion{EmojiType: elem.EmojiType})
			case "hr":
				lineElements = append(lineElements, &message.RichTextHr{})
			case "code_block":
				lineElements = append(lineElements, &message.RichTextCodeBlock{Text: elem.Text, Language: elem.Language})
			case "md":
				lineElements = append(lineElements, &message.RichTextMd{Text: elem.Text})
			}
		}
		if len(lineElements) > 0 {
			builder.AddLine(lineElements...)
		}
	}

	// 添加"已收到"的新行
	builder.AddText("\n已收到")

	// 回复富文本消息
	_, err := b.msgSender.ReplyMessage(ctx, messageID, builder)
	if err != nil {
		b.logger.Error("Failed to reply rich text message", zap.Error(err))
		return err
	}

	b.logger.Info("Rich text message replied successfully", zap.String("message_id", messageID))
	return nil
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

	b.logger.Info("Received reaction created event",
		zap.String("message_id", *reaction.MessageId),
		zap.String("emoji_type", emojiType),
		zap.String("user_id", userID),
	)

	// 同样的表情反应回去
	_, err := b.msgSender.AddReaction(ctx, *reaction.MessageId, message.EmojiType(emojiType))
	if err != nil {
		b.logger.Error("Failed to add reaction", zap.Error(err))
		return err
	}

	b.logger.Info("Reaction added successfully",
		zap.String("message_id", *reaction.MessageId),
		zap.String("emoji_type", emojiType),
	)

	return nil
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

	b.logger.Info("Received reaction deleted event",
		zap.String("message_id", *reaction.MessageId),
		zap.String("emoji_type", emojiType),
		zap.String("user_id", userID),
	)

	// 这里可以添加删除表情反应的逻辑，但通常不需要

	return nil
}

// replyText 回复文本消息
func (b *EchoBot) replyText(ctx context.Context, messageID, text string) error {
	builder := message.NewTextMessageBuilder(text)
	_, err := b.msgSender.ReplyMessage(ctx, messageID, builder)
	if err != nil {
		b.logger.Error("Failed to reply message", zap.Error(err))
		return err
	}
	b.logger.Info("Message replied successfully", zap.String("message_id", messageID), zap.String("text", text))
	return nil
}
