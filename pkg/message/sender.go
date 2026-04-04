package message

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"
)

// DefaultMessageSender 默认的消息发送器实现
type DefaultMessageSender struct {
	client *lark.Client
	logger *zap.Logger
}

// NewMessageSender 创建消息发送器
func NewMessageSender(client *lark.Client, logger *zap.Logger) MessageSender {
	return &DefaultMessageSender{
		client: client,
		logger: logger,
	}
}

// SendMessage 发送消息
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

	s.logger.Info("Message sent successfully",
		zap.String("receive_id_type", string(receiveIDType)),
		zap.String("receive_id", receiveID),
		zap.String("message_type", builder.MessageType()),
	)
	return resp, nil
}

// ReplyMessage 回复消息
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

	s.logger.Info("Message replied successfully",
		zap.String("message_id", messageID),
		zap.String("message_type", builder.MessageType()),
	)
	return resp, nil
}

// AddReaction 添加表情回复
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

	s.logger.Info("Reaction added successfully",
		zap.String("message_id", messageID),
		zap.String("emoji_type", string(emojiType)),
	)
	return resp, nil
}

// 确保 DefaultMessageSender 实现了 MessageSender 接口
var _ MessageSender = (*DefaultMessageSender)(nil)
