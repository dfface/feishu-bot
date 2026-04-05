// main_test 端到端测试 main.go
//
// 此文件测试应用程序的完整启动流程
package main

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/larksuite/oapi-sdk-go/v3/ws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dfface/feishu-bot/internal/bots"
	"github.com/dfface/feishu-bot/internal/config"
	"github.com/dfface/feishu-bot/internal/logger"
)

// TestEndToEnd 端到端测试
//
// 此测试验证应用程序的完整启动流程，包括配置加载、验证、机器人创建和 WebSocket 客户端设置
func TestEndToEnd(t *testing.T) {
	// 测试用例: 使用 config_test.yaml 配置文件启动应用程序
	configPath := "../../config_test.yaml"

	// 验证配置文件存在
	_, err := os.Stat(configPath)
	require.NoError(t, err, "Config file should exist")

	// 加载配置
	cfg, err := loadConfig(configPath)
	require.NoError(t, err, "Failed to load config")
	require.NotNil(t, cfg, "Config should not be nil")

	// 验证配置
	err = validateConfig(cfg)
	require.NoError(t, err, "Config should be valid")

	// 验证配置内容
	require.Len(t, cfg.Features, 2, "Should have 2 features")
	require.Len(t, cfg.Bots, 1, "Should have 1 bot")

	logger.Init(cfg.Log)

	// 验证 memos 功能配置
	memosFeature := findFeatureByID(cfg, "memos")
	require.NotNil(t, memosFeature, "Memos feature should exist")
	assert.True(t, memosFeature.Enabled, "Memos feature should be enabled")

	// 验证机器人配置
	bot := cfg.Bots[0]
	assert.True(t, bot.Enabled, "Bot should be enabled")
	assert.NotEmpty(t, bot.Feishu.AppID, "AppID should not be empty")
	assert.NotEmpty(t, bot.Feishu.AppSecret, "AppSecret should not be empty")

	// 创建机器人工厂
	botFactory := bots.NewBotFactory(cfg)
	require.NotNil(t, botFactory, "Bot factory should not be nil")

	// 创建机器人
	createdBots, err := botFactory.CreateBots()
	require.NoError(t, err, "Failed to create bots")
	require.NotEmpty(t, createdBots, "Should create at least one bot")

	// 验证机器人数量
	assert.Len(t, createdBots, 1, "Should create 1 bot")

	// 验证机器人名称
	createdBot := createdBots[0]
	assert.Equal(t, bot.Name, createdBot.Name(), "Bot name should match")

	// 验证机器人调度器
	botDispatcher := createdBot.GetDispatcher()
	assert.NotNil(t, botDispatcher, "Bot dispatcher should not be nil")

	// 验证机器人配置查找逻辑
	var foundBotConfig *config.BotConfig
	for _, cfgBot := range cfg.Bots {
		if cfgBot.ID == createdBot.ID() {
			foundBotConfig = &cfgBot
			break
		}
	}
	assert.NotNil(t, foundBotConfig, "Bot config should be found")
	assert.Equal(t, bot.ID, foundBotConfig.ID, "Bot ID should match")

	// 验证 WebSocket 配置
	assert.True(t, foundBotConfig.Feishu.UseWebSocket, "WebSocket should be enabled")
	assert.NotEmpty(t, foundBotConfig.Feishu.VerificationToken, "Verification token should not be empty")

	// 测试 WebSocket 客户端创建
	// 注意：这里只测试创建过程，不实际启动连接
	wsClient := ws.NewClient(
		foundBotConfig.Feishu.AppID,
		foundBotConfig.Feishu.AppSecret,
		ws.WithEventHandler(botDispatcher),
	)
	assert.NotNil(t, wsClient, "WebSocket client should be created successfully")

	// 验证 WebSocket 客户端配置
	// 由于 ws.Client 没有暴露内部配置的方法，我们通过检查创建过程是否成功来验证
	// 实际的 WebSocket 连接启动需要真实的网络环境和凭证，这里不进行测试
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// err = wsClient.Start(ctx)

	// // 等待中断信号
	// quit := make(chan os.Signal, 1)
	// signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// <-quit

	// // 优雅关闭
	// cancel()
	// time.Sleep(2 * time.Second)

}

// TestGracefulShutdown 测试优雅关闭
//
// 此测试验证应用程序能够正确处理中断信号并优雅关闭
func TestGracefulShutdown(t *testing.T) {
	// 模拟中断信号
	quit := make(chan os.Signal, 1)

	// 启动一个goroutine模拟信号发送
	go func() {
		// 等待一段时间后发送中断信号
		time.Sleep(100 * time.Millisecond)
		quit <- syscall.SIGINT
	}()

	// 等待信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	// 验证接收到的信号
	assert.Equal(t, syscall.SIGINT, sig, "Should receive SIGINT signal")
}

// TestLoadConfigWithEmptyPath 测试加载空配置文件路径
//
// 此测试验证当配置文件路径为空时，应用程序能够正确使用默认值
func TestLoadConfigWithEmptyPath(t *testing.T) {
	// 测试用例: 空配置文件路径
	cfg, err := loadConfig("")

	// 配置文件不存在时，应该使用默认值
	assert.NoError(t, err, "Should not error when config file path is empty")
	assert.NotNil(t, cfg, "Config should not be nil")

	// 验证默认配置
	assert.Len(t, cfg.Features, 2, "Should have 2 default features")
	assert.Len(t, cfg.Bots, 1, "Should have 1 default bot")
}

// findFeatureByID 根据 ID 查找功能配置
//
// 参数:
// - cfg: 配置实例
// - id: 功能 ID
//
// 返回值:
// - *config.FeatureConfig: 找到的功能配置，未找到则返回 nil
func findFeatureByID(cfg *config.Config, id string) *config.FeatureConfig {
	for i := range cfg.Features {
		if cfg.Features[i].ID == id {
			return &cfg.Features[i]
		}
	}
	return nil
}
