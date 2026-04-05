// bots 机器人实现包
//
// 此包实现了具体的机器人功能，包括功能管理和消息处理。
// 主要包含：
// 1. Bot 结构体：统一的机器人实现
// 2. 功能注册和管理方法
// 3. 消息处理和路由方法
package bots

import (
	"context"
	"fmt"
	"strings"
	"time"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"

	"github.com/dfface/feishu-bot/internal/bot"
	"github.com/dfface/feishu-bot/internal/features"
	"github.com/dfface/feishu-bot/internal/logger"
	"github.com/dfface/feishu-bot/internal/message"
)

// Bot 统一的机器人实现
// 实现 Bot 接口，提供功能管理和消息处理能力
//
// 此结构体是机器人的核心实现，通过嵌入 bot.BaseBot 获得基础功能，
// 并添加了功能管理和消息路由的能力。
// Bot 支持注册多个功能，并根据消息前缀或默认设置路由到相应的功能。
type Bot struct {
	*bot.BaseBot                                  // 基础机器人实例，提供消息处理和发送能力
	features          map[string]features.Feature // 功能映射表，键为功能 ID，值为功能实例
	defaultFeatureID  string                      // 默认功能 ID，当没有匹配的功能时使用
	processedMessages map[string]time.Time        // 已处理的消息 ID 集合，值为处理时间
}

const (
	// DefaultRetentionDuration 默认保留时间，3秒内消息
	DefaultIdempotencyInterval = 60 * time.Second
	DefaultRetentionInterval   = 120 * time.Second
)

// NewBot 创建机器人
//
// 此函数创建一个新的机器人实例，初始化基础机器人和功能映射表。
//
// 参数：
// - id：机器人唯一标识符
// - name：机器人名称
// - description：机器人描述
// - client：飞书 API 客户端
//
// 返回值：
// - *Bot：创建的机器人实例
func NewBot(id, name, description string, client *lark.Client) *Bot {
	return &Bot{
		BaseBot:           bot.NewBaseBot(id, name, description, client),
		features:          make(map[string]features.Feature),
		processedMessages: make(map[string]time.Time),
	}
}

// RegisterFeature 注册功能
//
// 此方法将功能注册到机器人，使其可以被调用。
// 注册时会自动调用功能的 SetBaseBot 方法，为功能提供访问机器人基础功能的能力。
//
// 参数：
// - feature：要注册的功能实例
func (b *Bot) RegisterFeature(feature features.Feature) {
	b.features[feature.ID()] = feature
	feature.SetBaseBot(b.BaseBot)
}

// SetDefaultFeature 设置默认功能
//
// 此方法设置默认功能 ID，当消息不匹配任何功能前缀时使用。
//
// 参数：
// - featureID：默认功能的 ID
func (b *Bot) SetDefaultFeature(featureID string) {
	b.defaultFeatureID = featureID
}

// cleanupProcessedMessages 清理过期的消息 ID
// 只保留最近 3 秒内的消息 ID
func (b *Bot) cleanupProcessedMessages() {
	retentionDuration := DefaultRetentionInterval // 保留最近 3 秒的消息 ID
	cutoffTime := time.Now().Add(-retentionDuration)

	for msgID, processTime := range b.processedMessages {
		if processTime.Before(cutoffTime) {
			delete(b.processedMessages, msgID)
		}
	}
}

// HandleMessage 处理消息
//
// 此方法是消息处理的核心，负责接收消息并路由到相应的功能。
// 主要完成以下工作：
// 1. 解析消息内容
// 2. 定期清理过期的消息 ID
// 3. 幂等处理：检查消息是否已经被处理过（3秒内的相同消息不处理）
// 4. 根据消息前缀匹配功能，如匹配到则清除前缀
// 5. 如果没有匹配的功能，使用默认功能
// 6. 调用功能的处理方法
//
// 参数：
// - ctx：上下文，用于控制请求的生命周期
// - event：消息事件，包含消息内容和发送者信息
//
// 返回值：
// - error：处理过程中的错误，成功则返回 nil
func (b *Bot) HandleMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	sender := event.Event.Sender

	// 解析消息内容
	msgContent, err := b.ProcessMessage(ctx, event)
	if err != nil {
		logger.Error("Failed to process message", zap.Error(err))
		return b.SendText(ctx, *sender.SenderId.OpenId, "消息处理失败")
	}

	// 定期清理过期的消息 ID
	if len(b.processedMessages)%10 == 0 { // 每处理 10 条消息清理一次
		b.cleanupProcessedMessages()
	}

	// 幂等处理：检查消息是否已经被处理过（3秒内的相同消息不处理）
	if processTime, exists := b.processedMessages[msgContent.ID]; exists {
		if time.Since(processTime) < DefaultIdempotencyInterval {
			logger.Info("Message already processed within 3 seconds", zap.String("message_id", msgContent.ID))
			return nil
		}
	}

	// 标记消息为已处理
	b.processedMessages[msgContent.ID] = time.Now()

	var matchedFeature features.Feature

	// 查找匹配的功能
	if msgContent.Type == message.MessageTypeText {
		text := msgContent.Text
		matchedFeature, text = b.matchFeature(text)
		msgContent.Text = text
	} else if msgContent.Type == message.MessageTypePost {
		// 富文本消息的规则是：第一行得有功能前缀
		if msgContent.RichText == nil || len(msgContent.RichText.Content) == 0 || len(msgContent.RichText.Content[0]) == 0 {
			logger.Error("Rich text message err", zap.Any("rich_text", msgContent))
			return b.SendText(ctx, *sender.SenderId.OpenId, "富文本消息解析失败")
		}
		text := msgContent.RichText.Content[0][0].Text
		matchedFeature, text = b.matchFeature(text)
		msgContent.RichText.Content[0][0].Text = text
	} else {
		// 其他消息类型都有摘要，按文本消息的处理方式来，比如文件有文件名的；但是图片被飞书自己重命名了
		text := msgContent.Text
		matchedFeature, text = b.matchFeature(text)
		msgContent.Text = text
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
	return matchedFeature.HandleMessage(ctx, msgContent)
}

// HandleP2PChatEntered 处理机器人进入单聊事件
//
// 此方法在机器人被添加到单聊时调用，发送欢迎消息给用户，
// 包含机器人的基本信息和可用功能列表。
//
// 参数：
// - ctx：上下文，用于控制请求的生命周期
// - event：单聊进入事件，包含用户信息
//
// 返回值：
// - error：处理过程中的错误，成功则返回 nil
func (b *Bot) HandleP2PChatEntered(ctx context.Context, event *larkim.P2ChatAccessEventBotP2pChatEnteredV1) error {
	// 获取用户的 open_id
	openID := *event.Event.OperatorId.OpenId

	// 构建欢迎消息（富文本）
	builder := message.NewRichTextMessageBuilder()

	// 添加标题和问候
	builder.SetTitle("你好！我是 " + b.Name()).NewParagraph().NewLine()
	builder.AddText(b.Description())

	// 添加功能列表标题
	builder.AddBoldText("我提供以下功能：").NewParagraph()

	// 添加功能列表
	for _, feature := range b.features {
		builder.AddMd(fmt.Sprintf("- **%s**(%s): %s", feature.Name(), feature.MatchPrefix(), feature.Description()))
	}

	// 添加使用方法
	builder.AddBoldText("使用方法：").NewParagraph()
	builder.AddMd("- 在消息中以功能的前缀开头，例如：`!echo 你好`")
	if b.defaultFeatureID != "" {
		builder.AddMd("- 如果你不添加前缀，那么将使用默认功能：" + "**" + b.features[b.defaultFeatureID].Name() + "**")
	}

	// 发送欢迎消息
	return b.SendRichText(ctx, openID, builder)
}

// matchFeature 匹配功能并更新消息内容
//
// 此方法根据文本内容匹配功能，并更新消息内容（移除前缀）。
func (b *Bot) matchFeature(text string) (features.Feature, string) {
	for _, feature := range b.features {
		if strings.HasPrefix(text, feature.MatchPrefix()) {
			// 移除前缀
			text = strings.TrimPrefix(text, feature.MatchPrefix())
			text = strings.TrimSpace(text)
			return feature, text
		}
	}
	return nil, text
}

func (b *Bot) HandleP2PChatEnteredReturnEmpty(ctx context.Context, event *larkim.P2ChatAccessEventBotP2pChatEnteredV1) error {
	return nil
}
