package bots

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"

	"github.com/dfface/feishu-bot/pkg/bot"
	"github.com/dfface/feishu-bot/pkg/config"
	"github.com/dfface/feishu-bot/pkg/message"
	"github.com/dfface/feishu-bot/pkg/memos"
)

// MemosBot Memos 机器人实现
type MemosBot struct {
	*bot.BaseBot
	memosClient  *memos.Client
	msgProcessor *message.Processor
	cfg          *config.Config
	logger       *zap.Logger
	tempDir      string
}

// NewMemosBot 创建 Memos 机器人
func NewMemosBot(name string, client *lark.Client, memosClient *memos.Client, cfg *config.Config, logger *zap.Logger) *MemosBot {
	tempDir := filepath.Join(os.TempDir(), "feishu-bot-memos")
	_ = os.MkdirAll(tempDir, 0755)

	baseBot := bot.NewBaseBot(name, client, cfg.Feishu.VerificationToken, cfg.Feishu.EncryptKey)

	bot := &MemosBot{
		BaseBot:      baseBot,
		memosClient:  memosClient,
		msgProcessor: message.NewProcessor(client, logger),
		cfg:          cfg,
		logger:       logger,
		tempDir:      tempDir,
	}

	// 设置事件处理器
	bot.GetDispatcher().OnP2MessageReceiveV1(func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
		return bot.HandleMessage(ctx, event)
	})

	return bot
}

// HandleMessage 处理消息事件
func (b *MemosBot) HandleMessage(ctx context.Context, event interface{}) error {
	imEvent := event.(*larkim.P2MessageReceiveV1)
	msg := imEvent.Event.Message
	sender := imEvent.Event.Sender

	b.logger.Info("Processing message",
		zap.String("message_id", *msg.MessageId),
		zap.String("sender_id", *sender.SenderId.OpenId),
		zap.String("chat_type", *msg.ChatType),
		zap.String("message_type", *msg.MessageType),
	)

	// 使用消息处理器解析消息并自动下载资源
	msgContent, err := b.msgProcessor.Process(ctx, msg)
	if err != nil {
		b.logger.Error("Failed to process message", zap.Error(err))
		return b.replyText(ctx, *sender.SenderId.OpenId, fmt.Sprintf("消息处理失败: %v", err))
	}

	// 处理文本内容
	content := msgContent.Text

	// 第一步：创建 Memo
	visibility := memos.Visibility(b.cfg.Memos.DefaultVisibility)
	memo, err := b.memosClient.CreateMemo(ctx, content, visibility)
	if err != nil {
		b.logger.Error("Failed to create memo", zap.Error(err))
		return b.replyText(ctx, *sender.SenderId.OpenId, fmt.Sprintf("保存失败: %v", err))
	}

	// 第二步：处理已下载的资源，上传到 Memos 并关联到 Memo
	for _, res := range msgContent.Resources {
		if !res.Downloaded {
			b.logger.Warn("Resource not downloaded", zap.String("type", res.Type))
			continue
		}

		// 上传到 Memos
		attachment, err := b.memosClient.UploadResource(ctx, res.LocalPath, memo.Name)
		if err != nil {
			b.logger.Error("Failed to upload resource to memos", zap.Error(err), zap.String("path", res.LocalPath))
			continue
		}
		b.logger.Info("Resource uploaded to memos", 
			zap.String("attachment_name", attachment.Name), 
			zap.String("filename", attachment.Filename),
		)


		// 清理本地文件
		_ = os.Remove(res.LocalPath)
	}

	// 回复成功
	return b.replyText(ctx, *sender.SenderId.OpenId, fmt.Sprintf("已保存到 Memos"))
}

// HandleCardAction 处理卡片交互事件
func (b *MemosBot) HandleCardAction(ctx context.Context, event interface{}) error {
	// 暂不处理卡片交互
	b.logger.Info("Received card action", zap.Any("event", event))
	return nil
}



// replyText 回复文本消息
func (b *MemosBot) replyText(ctx context.Context, receiveID, text string) error {
	// 暂时注释掉回复功能，待修复
	b.logger.Info("Would send reply", zap.String("receive_id", receiveID), zap.String("text", text))
	return nil
}
