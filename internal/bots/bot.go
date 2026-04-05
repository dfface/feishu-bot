package bots

import (
	"context"
	"strings"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"

	"github.com/dfface/feishu-bot/internal/features"
	"github.com/dfface/feishu-bot/pkg/bot"
)

// Bot 统一的机器人实现
type Bot struct {
	*bot.BaseBot
	features         map[string]features.Feature
	defaultFeatureID string
}

// NewBot 创建机器人
func NewBot(name string, client *lark.Client, logger *zap.Logger) *Bot {
	return &Bot{
		BaseBot:  bot.NewBaseBot(name, client, logger),
		features: make(map[string]features.Feature),
	}
}

// RegisterFeature 注册功能
func (b *Bot) RegisterFeature(feature features.Feature) {
	b.features[feature.ID()] = feature
	if f, ok := feature.(interface{ SetBaseBot(*bot.BaseBot) }); ok {
		f.SetBaseBot(b.BaseBot)
	}
}

// SetDefaultFeature 设置默认功能
func (b *Bot) SetDefaultFeature(featureID string) {
	b.defaultFeatureID = featureID
}

// HandleMessage 处理消息
func (b *Bot) HandleMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	msg := event.Event.Message
	sender := event.Event.Sender

	// 解析消息内容
	msgContent, err := b.MsgProcessor.Process(ctx, msg)
	if err != nil {
		b.Logger.Error("Failed to process message", zap.Error(err))
		return b.SendText(ctx, *sender.SenderId.OpenId, "消息处理失败")
	}

	// 查找匹配的功能
	var matchedFeature features.Feature
	text := msgContent.Text

	// 首先尝试根据前缀匹配功能
	for _, feature := range b.features {
		if strings.HasPrefix(text, feature.MatchPrefix()) {
			matchedFeature = feature
			break
		}
	}

	// 如果没有匹配的功能，使用默认功能
	if matchedFeature == nil && b.defaultFeatureID != "" {
		matchedFeature = b.features[b.defaultFeatureID]
	}

	// 如果还是没有找到功能，返回错误
	if matchedFeature == nil {
		return b.SendText(ctx, *sender.SenderId.OpenId, "没有可用的功能")
	}

	// 调用功能的处理方法
	return matchedFeature.HandleMessage(ctx, event)
}
