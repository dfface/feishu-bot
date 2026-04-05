// bots 机器人工厂实现包
//
// 此包实现了机器人工厂，负责根据配置创建和初始化机器人。
// 主要包含：
// 1. BotFactory 结构体：机器人工厂
// 2. CreateBots 方法：创建机器人
// 3. createBot 方法：创建单个机器人
package bots

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/dfface/feishu-bot/internal/bot"
	"github.com/dfface/feishu-bot/internal/config"
	"github.com/dfface/feishu-bot/internal/features"
	"github.com/dfface/feishu-bot/internal/logger"
	lark "github.com/larksuite/oapi-sdk-go/v3"
)

// BotFactory 机器人工厂
// 负责根据配置创建和初始化机器人
//
// 工厂模式是创建对象的经典设计模式，它封装了对象的创建逻辑。
// BotFactory 负责读取配置、创建机器人实例、注册功能和设置消息处理器。
type BotFactory struct {
	config          *config.Config            // 应用配置
	featureRegistry *features.FeatureRegistry // 功能注册中心
}

// NewBotFactory 创建机器人工厂
//
// 此函数创建一个新的机器人工厂实例，并初始化功能注册中心。
// 主要完成以下工作：
// 1. 创建功能注册中心
// 2. 注册所有内置功能
//
// 参数：
// - cfg：应用配置
//
// 返回值：
// - *BotFactory：创建的机器人工厂实例
func NewBotFactory(cfg *config.Config) *BotFactory {
	registry := features.NewFeatureRegistry()
	features.RegisterFeatures(registry)

	return &BotFactory{
		config:          cfg,
		featureRegistry: registry,
	}
}

// CreateBots 根据配置创建机器人
//
// 此方法遍历配置中的所有机器人，创建启用的机器人实例。
// 主要完成以下工作：
// 1. 遍历配置中的机器人列表
// 2. 跳过未启用的机器人
// 3. 创建机器人实例
// 4. 收集创建成功的机器人
//
// 返回值：
// - []bot.Bot：创建的机器人列表
// - error：创建过程中的错误，成功则返回 nil
func (f *BotFactory) CreateBots() ([]bot.Bot, error) {
	var createdBots []bot.Bot

	for _, botConfig := range f.config.Bots {
		if !botConfig.Enabled {
			continue
		}

		createdBot, err := f.createBot(botConfig)
		if err != nil {
			logger.Error("Failed to create bot",
				zap.String("bot_id", botConfig.ID),
				zap.Error(err))
			continue
		}

		createdBots = append(createdBots, createdBot)
	}

	return createdBots, nil
}

// createBot 创建机器人
//
// 此方法创建单个机器人实例，并完成初始化。
// 主要完成以下工作：
// 1. 验证机器人配置（至少有一个功能）
// 2. 创建飞书客户端
// 3. 创建机器人实例
// 4. 注册和初始化功能
// 5. 设置默认功能
// 6. 设置消息处理器
//
// 参数：
// - botConfig：机器人配置
//
// 返回值：
// - bot.Bot：创建的机器人实例
// - error：创建过程中的错误，成功则返回 nil
func (f *BotFactory) createBot(botConfig config.BotConfig) (bot.Bot, error) {
	if len(botConfig.Features) == 0 {
		return nil, fmt.Errorf("bot must have at least one feature")
	}

	// 为每个机器人创建独立的 lark.Client，使用机器人自己的飞书配置
	feishuClient := lark.NewClient(botConfig.Feishu.AppID, botConfig.Feishu.AppSecret)
	newBot := NewBot(botConfig.ID, botConfig.Name, botConfig.Description, feishuClient)

	// 注册功能
	for _, featureMapping := range botConfig.Features {
		// 首先根据用户自定义的 ID 查找 FeatureConfig
		var featureConfig *config.FeatureConfig
		for i, cfg := range f.config.Features {
			if cfg.ID == featureMapping.FeatureID {
				featureConfig = &f.config.Features[i]
				break
			}
		}

		// 如果找不到 FeatureConfig，跳过
		if featureConfig == nil {
			logger.Warn("Feature config not found", zap.String("feature_id", featureMapping.FeatureID))
			continue
		}

		// 然后根据 InternalID 查找功能实例
		feature := f.featureRegistry.Get(featureConfig.InternalID)
		if feature == nil {
			logger.Warn("Feature not found", zap.String("internal_id", featureConfig.InternalID))
			continue
		}

		// 初始化功能
		if featureConfig.Enabled {
			if err := feature.Initialize(featureConfig); err != nil {
				logger.Error("Failed to initialize feature",
					zap.String("internal_id", featureConfig.InternalID),
					zap.Error(err))
			}
		}

		// 注册功能
		newBot.RegisterFeature(feature)

		// 设置默认功能
		if featureMapping.Default {
			newBot.SetDefaultFeature(feature.ID())
		} else if len(f.config.Features) == 1 {
			newBot.SetDefaultFeature(feature.ID())
		}
	}

	// 设置消息处理器
	newBot.OnMessage(newBot.HandleMessage)

	return newBot, nil
}
