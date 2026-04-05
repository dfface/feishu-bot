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
	"strings"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"go.uber.org/zap"

	"github.com/dfface/feishu-bot/internal/bot"
	"github.com/dfface/feishu-bot/internal/features"
	"github.com/dfface/feishu-bot/internal/logger"
)

// Bot 统一的机器人实现
// 实现 Bot 接口，提供功能管理和消息处理能力
//
// 此结构体是机器人的核心实现，通过嵌入 bot.BaseBot 获得基础功能，
// 并添加了功能管理和消息路由的能力。
// Bot 支持注册多个功能，并根据消息前缀或默认设置路由到相应的功能。
type Bot struct {
	*bot.BaseBot                                 // 基础机器人实例，提供消息处理和发送能力
	features         map[string]features.Feature // 功能映射表，键为功能 ID，值为功能实例
	defaultFeatureID string                      // 默认功能 ID，当没有匹配的功能时使用
}

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
		BaseBot:  bot.NewBaseBot(id, name, description, client),
		features: make(map[string]features.Feature),
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

// HandleMessage 处理消息
//
// 此方法是消息处理的核心，负责接收消息并路由到相应的功能。
// 主要完成以下工作：
// 1. 解析消息内容
// 2. 根据消息前缀匹配功能
// 3. 如果没有匹配的功能，使用默认功能
// 4. 调用功能的处理方法
//
// 参数：
// - ctx：上下文，用于控制请求的生命周期
// - event：消息事件，包含消息内容和发送者信息
//
// 返回值：
// - error：处理过程中的错误，成功则返回 nil
func (b *Bot) HandleMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	msg := event.Event.Message
	sender := event.Event.Sender

	// 解析消息内容
	msgContent, err := b.MsgProcessor.Process(ctx, msg)
	if err != nil {
		logger.Error("Failed to process message", zap.Error(err))
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
