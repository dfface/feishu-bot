// main 飞书机器人应用程序入口
//
// 此包是飞书机器人应用程序的主入口，负责初始化和启动机器人。
// 主要完成以下工作：
// 1. 加载和验证配置
// 2. 初始化日志系统
// 3. 创建机器人工厂和机器人实例
// 4. 启动 WebSocket 客户端连接飞书服务器
// 5. 处理优雅关闭
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/larksuite/oapi-sdk-go/v3/ws"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/dfface/feishu-bot/internal/bots"
	"github.com/dfface/feishu-bot/internal/config"
)

// main 应用程序入口函数
//
// 此函数是应用程序的主入口，负责协调整个应用程序的启动流程。
// 主要完成以下工作：
// 1. 解析命令行参数
// 2. 加载和验证配置
// 3. 初始化日志系统
// 4. 创建机器人工厂和机器人实例
// 5. 为每个机器人创建 WebSocket 客户端
// 6. 启动 WebSocket 客户端
// 7. 等待中断信号
// 8. 优雅关闭
func main() {
	// 解析命令行参数
	configPath := flag.String("config", "", "Path to config file")
	flag.Parse()

	// 加载配置
	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 验证配置
	if err := validateConfig(cfg); err != nil {
		fmt.Printf("Failed to validate config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logger, err := initLogger(cfg.Log)
	if err != nil {
		fmt.Printf("Failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting feishu-bot...")

	// 创建机器人工厂
	botFactory := bots.NewBotFactory(cfg, logger)

	// 创建机器人
	bots, err := botFactory.CreateBots()
	if err != nil {
		logger.Fatal("Failed to create bots", zap.Error(err))
	}

	if len(bots) == 0 {
		logger.Fatal("No bots created")
	}

	// 为每个机器人创建一个 WebSocket 客户端
	var wsClients []*ws.Client
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i, b := range bots {
		botDispatcher := b.GetDispatcher()

		// 查找机器人配置，获取飞书配置
		var botConfig *config.BotConfig
		for _, cfgBot := range cfg.Bots {
			if cfgBot.Name == b.Name() {
				botConfig = &cfgBot
				break
			}
		}

		if botConfig == nil {
			logger.Fatal("Bot config not found", zap.String("bot_name", b.Name()))
		}

		// 创建 WebSocket 客户端，使用机器人自己的飞书配置
		wsClient := ws.NewClient(
			botConfig.Feishu.AppID,
			botConfig.Feishu.AppSecret,
			ws.WithEventHandler(botDispatcher),
		)

		wsClients = append(wsClients, wsClient)

		// 启动 WebSocket 客户端
		logger.Info("Starting WebSocket client for bot",
			zap.String("bot_name", b.Name()),
			zap.Int("client_index", i),
		)

		err = wsClient.Start(ctx)
		if err != nil {
			logger.Fatal("Failed to start WebSocket client",
				zap.String("bot_name", b.Name()),
				zap.Error(err),
			)
		}
	}

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down WebSocket clients...")

	// 优雅关闭
	cancel()
	time.Sleep(2 * time.Second)

	logger.Info("All WebSocket clients exited")
}

// loadConfig 加载配置
//
// 此函数从配置文件或环境变量加载应用程序配置。
// 主要完成以下工作：
// 1. 设置默认值
// 2. 从环境变量读取配置
// 3. 从配置文件读取配置
// 4. 解析配置到结构体
// 5. 设置默认功能和机器人配置（如果配置为空）
//
// 参数：
// - configPath：配置文件路径，如果为空则从默认路径查找
//
// 返回值：
// - *config.Config：加载的配置实例
// - error：加载过程中的错误，成功则返回 nil
func loadConfig(configPath string) (*config.Config, error) {
	v := viper.New()

	// 设置默认值
	setDefaults(v)

	// 从环境变量读取
	v.AutomaticEnv()
	v.SetEnvPrefix("FEISHU_BOT")

	// 从配置文件读取
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
	}

	// 读取配置
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// 配置文件不存在时使用环境变量
	}

	var cfg config.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 设置默认功能配置
	if len(cfg.Features) == 0 {
		cfg.Features = []config.FeatureConfig{
			{
				ID:          "echo",
				Name:        "回声功能",
				Enabled:     true,
				Description: "原样回复消息",
				Config: map[string]interface{}{
					"prefix": "!echo",
				},
			},
			{
				ID:          "memos",
				Name:        "Memos 保存",
				Enabled:     true,
				Description: "保存消息到 Memos",
				Config: map[string]interface{}{
					"prefix": "!memos",
				},
			},
		}
	}

	// 设置默认机器人配置
	if len(cfg.Bots) == 0 {
		cfg.Bots = []config.BotConfig{
			{
				ID:      "multi-bot",
				Name:    "多功能机器人",
				Enabled: true,
				Feishu: config.FeishuConfig{
					AppID:        "",
					AppSecret:    "",
					UseWebSocket: true,
				},
				Features: []config.FeatureMapping{
					{
						FeatureID: "echo",
						Default:   true,
					},
					{
						FeatureID: "memos",
					},
				},
				Config: map[string]interface{}{
					"welcome_message": "欢迎使用多功能机器人！",
				},
			},
		}
	}

	return &cfg, nil
}

// validateConfig 验证配置
//
// 此函数验证配置的完整性和正确性。
// 主要完成以下工作：
// 1. 检查所有启用的机器人的飞书配置是否完整
// 2. 检查所有启用的 Memos 功能的配置是否完整
//
// 参数：
// - cfg：配置实例
//
// 返回值：
// - error：验证过程中的错误，成功则返回 nil
func validateConfig(cfg *config.Config) error {
	// 检查所有启用的机器人的飞书配置
	for _, bot := range cfg.Bots {
		if bot.Enabled {
			if bot.Feishu.AppID == "" {
				return fmt.Errorf("feishu.app_id is required for bot: %s", bot.Name)
			}
			if bot.Feishu.AppSecret == "" {
				return fmt.Errorf("feishu.app_secret is required for bot: %s", bot.Name)
			}
		}
	}

	// 检查所有启用的 memos 功能的配置
	for _, feature := range cfg.Features {
		if feature.ID == "memos" && feature.Enabled {
			if feature.Config == nil {
				return fmt.Errorf("config is required for feature: %s", feature.Name)
			}
			memosConfig, ok := feature.Config["memos"].(map[string]interface{})
			if !ok {
				return fmt.Errorf("memos config is required for feature: %s", feature.Name)
			}
			if baseURL, ok := memosConfig["base_url"].(string); !ok || baseURL == "" {
				return fmt.Errorf("memos.base_url is required for feature: %s", feature.Name)
			}
			if accessToken, ok := memosConfig["access_token"].(string); !ok || accessToken == "" {
				return fmt.Errorf("memos.access_token is required for feature: %s", feature.Name)
			}
		}
	}

	return nil
}

// setDefaults 设置默认值
//
// 此函数设置配置的默认值，当配置文件或环境变量中没有指定时使用。
//
// 参数：
// - v：Viper 实例
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8080)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
}

// initLogger 初始化日志
//
// 此函数初始化日志系统，根据配置设置日志级别和格式。
// 主要完成以下工作：
// 1. 解析日志级别
// 2. 配置日志格式（json 或 console）
// 3. 配置日志输出路径
//
// 参数：
// - cfg：日志配置
//
// 返回值：
// - *zap.Logger：初始化的日志记录器
// - error：初始化过程中的错误，成功则返回 nil
func initLogger(cfg config.LogConfig) (*zap.Logger, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         cfg.Format,
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	if cfg.Format == "console" {
		config.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	return config.Build()
}
