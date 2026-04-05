// bot 包主要用于解决循环依赖问题 features 要依赖 bot
package bot

import (
	"context"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"

	"github.com/dfface/feishu-bot/internal/logger"
	"github.com/dfface/feishu-bot/internal/message"
)

// Bot 机器人接口
// 定义所有机器人必须实现的最小接口
type Bot interface {
	// Name 返回机器人名称
	Name() string
	// GetDispatcher 获取事件分发器，用于 WebSocket 事件处理
	GetDispatcher() *dispatcher.EventDispatcher
}

// BaseBot 基础机器人实现
// 提供所有机器人共有的功能和组件
type BaseBot struct {
	name         string
	client       *lark.Client
	dispatcher   *dispatcher.EventDispatcher
	MsgProcessor message.MessageReceiver // 消息处理器，用于解析接收到的消息
	MsgSender    message.MessageSender   // 消息发送器，用于发送和回复消息
	FileUploader message.FileUploader    // 文件上传器，用于上传图片和文件
}

// NewBaseBot 创建基础机器人
// 初始化所有机器人共有的组件
//
// 参数:
//
//	name - 机器人名称
//	client - 飞书 API 客户端
//
// 返回:
//
//	*BaseBot - 初始化好的基础机器人实例
func NewBaseBot(name string, client *lark.Client) *BaseBot {
	return &BaseBot{
		name:         name,
		client:       client,
		dispatcher:   dispatcher.NewEventDispatcher("", ""),
		MsgProcessor: message.NewProcessor(client),
		MsgSender:    message.NewMessageSender(client),
		FileUploader: message.NewFileUploader(client),
	}
}

// Name 返回机器人名称
func (b *BaseBot) Name() string {
	return b.name
}

// GetDispatcher 获取事件分发器
func (b *BaseBot) GetDispatcher() *dispatcher.EventDispatcher {
	return b.dispatcher
}

// GetClient 获取飞书客户端
func (b *BaseBot) GetClient() *lark.Client {
	return b.client
}

// ReplyText 便捷方法：回复文本消息
//
// 参数:
//
//	ctx - 上下文
//	messageID - 被回复的消息 ID
//	text - 回复的文本内容
//
// 返回:
//
//	error - 回复失败时返回错误
func (b *BaseBot) ReplyText(ctx context.Context, messageID string, text string) error {
	builder := message.NewTextMessageBuilder(text)
	_, err := b.MsgSender.ReplyMessage(ctx, messageID, builder)
	if err != nil {
		logger.Error("Failed to reply text message", zap.Error(err))
		return err
	}
	logger.Info("Text message replied successfully",
		zap.String("message_id", messageID),
		zap.String("text", text))
	return nil
}

// ReplyRichText 便捷方法：回复富文本消息
//
// 参数:
//
//	ctx - 上下文
//	messageID - 被回复的消息 ID
//	builder - 富文本消息构建器
//
// 返回:
//
//	error - 回复失败时返回错误
func (b *BaseBot) ReplyRichText(ctx context.Context, messageID string, builder *message.RichTextMessageBuilder) error {
	_, err := b.MsgSender.ReplyMessage(ctx, messageID, builder)
	if err != nil {
		logger.Error("Failed to reply rich text message", zap.Error(err))
		return err
	}
	logger.Info("Rich text message replied successfully",
		zap.String("message_id", messageID))
	return nil
}

// SendText 便捷方法：发送文本消息给用户
//
// 参数:
//
//	ctx - 上下文
//	receiveID - 接收者 ID（open_id）
//	text - 文本内容
//
// 返回:
//
//	error - 发送失败时返回错误
func (b *BaseBot) SendText(ctx context.Context, receiveID string, text string) error {
	builder := message.NewTextMessageBuilder(text)
	_, err := b.MsgSender.SendMessage(ctx, message.ReceiveIDTypeOpenID, receiveID, builder)
	if err != nil {
		logger.Error("Failed to send text message", zap.Error(err))
		return err
	}
	logger.Info("Text message sent successfully",
		zap.String("receive_id", receiveID),
		zap.String("text", text))
	return nil
}

// AddReaction 便捷方法：添加表情反应
//
// 参数:
//
//	ctx - 上下文
//	messageID - 消息 ID
//	emojiType - 表情类型
//
// 返回:
//
//	error - 添加失败时返回错误
func (b *BaseBot) AddReaction(ctx context.Context, messageID string, emojiType message.EmojiType) error {
	_, err := b.MsgSender.AddReaction(ctx, messageID, emojiType)
	if err != nil {
		logger.Error("Failed to add reaction", zap.Error(err))
		return err
	}
	logger.Info("Reaction added successfully",
		zap.String("message_id", messageID),
		zap.String("emoji_type", string(emojiType)))
	return nil
}

// OnMessage 便捷方法：注册消息接收事件处理器
//
// 参数:
//
//	handler - 事件处理函数
func (b *BaseBot) OnMessage(handler func(ctx context.Context, event *larkim.P2MessageReceiveV1) error) {
	b.dispatcher.OnP2MessageReceiveV1(handler)
}

// OnReactionCreated 便捷方法：注册表情反应创建事件处理器
//
// 参数:
//
//	handler - 事件处理函数
func (b *BaseBot) OnReactionCreated(handler func(ctx context.Context, event *larkim.P2MessageReactionCreatedV1) error) {
	b.dispatcher.OnP2MessageReactionCreatedV1(handler)
}

// OnReactionDeleted 便捷方法：注册表情反应删除事件处理器
//
// 参数:
//
//	handler - 事件处理函数
func (b *BaseBot) OnReactionDeleted(handler func(ctx context.Context, event *larkim.P2MessageReactionDeletedV1) error) {
	b.dispatcher.OnP2MessageReactionDeletedV1(handler)
}
