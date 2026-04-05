package bots

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/dfface/feishu-bot/internal/features"
	"github.com/dfface/feishu-bot/pkg/bot"
	"github.com/dfface/feishu-bot/pkg/config"
	lark "github.com/larksuite/oapi-sdk-go/v3"
)

// BotFactory 机器人工厂
type BotFactory struct {
	config          *config.Config
	logger          *zap.Logger
	featureRegistry *features.FeatureRegistry
}

// NewBotFactory 创建机器人工厂
func NewBotFactory(cfg *config.Config, logger *zap.Logger) *BotFactory {
	registry := features.NewFeatureRegistry()
	features.RegisterFeatures(registry)

	return &BotFactory{
		config:          cfg,
		logger:          logger,
		featureRegistry: registry,
	}
}

// CreateBots 根据配置创建机器人
func (f *BotFactory) CreateBots() ([]bot.Bot, error) {
	var createdBots []bot.Bot

	for _, botConfig := range f.config.Bots {
		if !botConfig.Enabled {
			continue
		}

		createdBot, err := f.createBot(botConfig)
		if err != nil {
			f.logger.Error("Failed to create bot",
				zap.String("bot_id", botConfig.ID),
				zap.Error(err))
			continue
		}

		createdBots = append(createdBots, createdBot)
	}

	return createdBots, nil
}

// createBot 创建机器人
func (f *BotFactory) createBot(botConfig config.BotConfig) (bot.Bot, error) {
	if len(botConfig.Features) == 0 {
		return nil, fmt.Errorf("bot must have at least one feature")
	}

	// 为每个机器人创建独立的 lark.Client，使用机器人自己的飞书配置
	feishuClient := lark.NewClient(botConfig.Feishu.AppID, botConfig.Feishu.AppSecret)
	newBot := NewBot(botConfig.Name, feishuClient, f.logger)

	// 注册功能
	for _, featureMapping := range botConfig.Features {
		feature := f.featureRegistry.Get(featureMapping.FeatureID)
		if feature == nil {
			f.logger.Warn("Feature not found", zap.String("feature_id", featureMapping.FeatureID))
			continue
		}

		// 初始化功能
		for _, featureConfig := range f.config.Features {
			if featureConfig.ID == featureMapping.FeatureID && featureConfig.Enabled {
				if err := feature.Initialize(&featureConfig); err != nil {
					f.logger.Error("Failed to initialize feature",
						zap.String("feature_id", featureMapping.FeatureID),
						zap.Error(err))
				}
				break
			}
		}

		// 注册功能
		newBot.RegisterFeature(feature)

		// 设置默认功能
		if featureMapping.Default {
			newBot.SetDefaultFeature(feature.ID())
		}
	}

	// 设置消息处理器
	newBot.OnMessage(newBot.HandleMessage)

	return newBot, nil
}
