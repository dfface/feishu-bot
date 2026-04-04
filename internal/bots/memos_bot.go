package bots

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"

	"github.com/dfface/feishu-bot/pkg/bot"
	"github.com/dfface/feishu-bot/pkg/config"
	"github.com/dfface/feishu-bot/pkg/converter"
	"github.com/dfface/feishu-bot/third_party/memos"
)

// MemosBot Memos 机器人 - 将飞书消息保存到 Memos
type MemosBot struct {
	*bot.BaseBot
	memosClient *memos.Client
	cfg         *config.Config
	tempDir     string
}

// NewMemosBot 创建 Memos 机器人
//
// 参数:
//   name - 机器人名称
//   client - 飞书 API 客户端
//   memosClient - Memos 客户端
//   cfg - 应用配置
//   logger - 日志记录器
//
// 返回:
//   *MemosBot - 初始化好的 Memos 机器人实例
func NewMemosBot(name string, client *lark.Client, memosClient *memos.Client, cfg *config.Config, logger *zap.Logger) *MemosBot {
	tempDir := filepath.Join(os.TempDir(), "feishu-bot-memos")
	_ = os.MkdirAll(tempDir, 0755)

	b := &MemosBot{
		BaseBot:     bot.NewBaseBot(name, client, logger),
		memosClient: memosClient,
		cfg:         cfg,
		tempDir:     tempDir,
	}

	// 设置事件处理器
	b.OnMessage(b.HandleMessage)

	return b
}

// HandleMessage 处理消息事件
func (b *MemosBot) HandleMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	msg := event.Event.Message
	sender := event.Event.Sender

	b.Logger.Info("Processing message",
		zap.String("message_id", *msg.MessageId),
		zap.String("sender_id", *sender.SenderId.OpenId),
		zap.String("chat_type", *msg.ChatType),
		zap.String("message_type", *msg.MessageType),
	)

	// 使用消息处理器解析消息并自动下载资源
	msgContent, err := b.MsgProcessor.Process(ctx, msg)
	if err != nil {
		b.Logger.Error("Failed to process message", zap.Error(err))
		return b.ReplyText(ctx, *sender.SenderId.OpenId, fmt.Sprintf("消息处理失败: %v", err))
	}

	// 使用转换器将 MessageContent 转换为 Memos 格式
	content, filePaths, err := converter.NewMemosConverter().ConvertMessageContent(msgContent)
	if err != nil {
		b.Logger.Error("Failed to convert message content", zap.Error(err))
		return b.ReplyText(ctx, *sender.SenderId.OpenId, fmt.Sprintf("消息转换失败: %v", err))
	}

	b.Logger.Info("Converted message content",
		zap.String("content", content),
		zap.Int("resource_count", len(filePaths)))

	// 创建带资源的 Memo
	visibility := memos.Visibility(b.cfg.Memos.DefaultVisibility)
	memo, attachments, err := b.memosClient.CreateMemoWithResources(ctx, content, visibility, filePaths)
	if err != nil {
		b.Logger.Error("Failed to create memo with resources", zap.Error(err))
		return b.ReplyText(ctx, *sender.SenderId.OpenId, fmt.Sprintf("保存失败: %v", err))
	}

	// 清理本地文件
	for _, path := range filePaths {
		_ = os.Remove(path)
	}

	b.Logger.Info("Memo created successfully",
		zap.String("memo_name", memo.Name),
		zap.Int("attachment_count", len(attachments)))

	// 回复成功
	return b.ReplyText(ctx, *sender.SenderId.OpenId, fmt.Sprintf("已保存到 Memos"))
}
