package message

import (
	"context"
	"fmt"

	"github.com/dfface/feishu-bot/internal/logger"
	"github.com/google/uuid"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"
)

// DefaultMessageSender 默认的消息发送器实现
// 实现了 MessageSender 接口，提供消息发送、回复和表情反应功能
type DefaultMessageSender struct {
	client *lark.Client
}

// NewMessageSender 创建消息发送器
//
// 参数:
//
//	client - 飞书 API 客户端
//
// 返回:
//
//	MessageSender - 初始化好的消息发送器实例
func NewMessageSender(client *lark.Client) MessageSender {
	return &DefaultMessageSender{
		client: client,
	}
}

// SendMessage 发送消息
// 向指定的接收者发送消息，支持多种接收者 ID 类型
//
// 参数:
//
//	ctx - 上下文，用于取消操作
//	receiveIDType - 接收者 ID 类型（open_id、user_id、union_id、email、chat_id）
//	receiveID - 接收者 ID
//	builder - 消息构建器，用于构建消息内容
//
// 返回:
//
//	*larkim.CreateMessageResp - 飞书 API 返回的响应
//	error - 发送失败时返回错误
func (s *DefaultMessageSender) SendMessage(ctx context.Context, receiveIDType ReceiveIDType, receiveID string, builder MessageBuilder) (*larkim.CreateMessageResp, error) {
	content, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build message: %w", err)
	}

	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(string(receiveIDType)).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			MsgType(builder.MessageType()).
			ReceiveId(receiveID).
			Content(content).
			Uuid(uuid.New().String()).
			Build()).
		Build()

	resp, err := s.client.Im.V1.Message.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	if !resp.Success() {
		return nil, fmt.Errorf("failed to send message: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	logger.Info("Message sent successfully",
		zap.String("receive_id_type", string(receiveIDType)),
		zap.String("receive_id", receiveID),
		zap.String("message_type", builder.MessageType()),
	)
	return resp, nil
}

// ReplyMessage 回复消息
// 回复指定的消息，支持各种消息类型
//
// 参数:
//
//	ctx - 上下文，用于取消操作
//	messageID - 被回复的消息 ID
//	builder - 消息构建器，用于构建回复消息内容
//
// 返回:
//
//	*larkim.ReplyMessageResp - 飞书 API 返回的响应
//	error - 回复失败时返回错误
func (s *DefaultMessageSender) ReplyMessage(ctx context.Context, messageID string, builder MessageBuilder) (*larkim.ReplyMessageResp, error) {
	content, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build message: %w", err)
	}

	req := larkim.NewReplyMessageReqBuilder().
		MessageId(messageID).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			MsgType(builder.MessageType()).
			Content(content).
			Uuid(uuid.New().String()).
			Build()).
		Build()

	resp, err := s.client.Im.V1.Message.Reply(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to reply message: %w", err)
	}

	if !resp.Success() {
		return nil, fmt.Errorf("failed to reply message: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	logger.Info("Message replied successfully",
		zap.String("message_id", messageID),
		zap.String("message_type", builder.MessageType()),
	)
	return resp, nil
}

// AddReaction 添加表情回复
// 为指定的消息添加表情反应
//
// 参数:
//
//	ctx - 上下文，用于取消操作
//	messageID - 要添加表情的消息 ID
//	emojiType - 表情类型，使用预定义的 EmojiType 常量
//
// 返回:
//
//	*larkim.CreateMessageReactionResp - 飞书 API 返回的响应
//	error - 添加失败时返回错误
func (s *DefaultMessageSender) AddReaction(ctx context.Context, messageID string, emojiType EmojiType) (*larkim.CreateMessageReactionResp, error) {
	resp, err := s.client.Im.V1.MessageReaction.Create(ctx,
		larkim.NewCreateMessageReactionReqBuilder().
			MessageId(messageID).
			Body(larkim.NewCreateMessageReactionReqBodyBuilder().
				ReactionType(larkim.NewEmojiBuilder().
					EmojiType(string(emojiType)).
					Build()).
				Build()).
			Build())

	if err != nil {
		return nil, fmt.Errorf("failed to add reaction: %w", err)
	}

	if !resp.Success() {
		return nil, fmt.Errorf("failed to add reaction: code=%d, msg=%s", resp.Code, resp.Msg)
	}

	logger.Info("Reaction added successfully",
		zap.String("message_id", messageID),
		zap.String("emoji_type", string(emojiType)),
	)
	return resp, nil
}

// 确保 DefaultMessageSender 实现了 MessageSender 接口
var _ MessageSender = (*DefaultMessageSender)(nil)
