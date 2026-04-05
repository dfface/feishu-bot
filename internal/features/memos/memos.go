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
type MemosFeature struct {
	id          string
	name        string
	description string
	prefix      string
	baseBot     *bot.BaseBot
	logger      *zap.Logger
	memosClient *memos.Client
	cfg         *config.Config
	memosConfig *config.MemosConfig
}

// NewMemosFeature 创建 Memos 功能
func NewMemosFeature() *MemosFeature {
	return &MemosFeature{
		id:          "memos",
		name:        "Memos 保存",
		description: "保存消息到 Memos",
		prefix:      "!memos",
	}
}

// ID 返回功能ID
func (f *MemosFeature) ID() string {
	return f.id
}

// Name 返回功能名称
func (f *MemosFeature) Name() string {
	return f.name
}

// Description 返回功能描述
func (f *MemosFeature) Description() string {
	return f.description
}

// MatchPrefix 返回匹配前缀
func (f *MemosFeature) MatchPrefix() string {
	return f.prefix
}

// Initialize 初始化功能
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
				f.logger,
			)
		}
	}

	return nil
}

// getStringValue 从 map 中获取字符串值
func getStringValue(m map[string]interface{}, key, defaultValue string) string {
	if value, ok := m[key].(string); ok {
		return value
	}
	return defaultValue
}

// SetBaseBot 设置基础机器人
func (f *MemosFeature) SetBaseBot(baseBot *bot.BaseBot) {
	f.baseBot = baseBot
	f.logger = baseBot.Logger
}

// SetMemosClient 设置 Memos 客户端
func (f *MemosFeature) SetMemosClient(client *memos.Client) {
	f.memosClient = client
}

// SetConfig 设置配置
func (f *MemosFeature) SetConfig(cfg *config.Config) {
	f.cfg = cfg
}

// SetMemosConfig 设置 Memos 配置
func (f *MemosFeature) SetMemosConfig(memosConfig *config.MemosConfig) {
	f.memosConfig = memosConfig
}

// HandleMessage 处理消息
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
