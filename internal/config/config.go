// config 配置管理包
//
// 此包负责应用程序的配置管理，包括配置的加载、解析和验证。
// 主要包含：
// 1. Config 结构体：应用程序的主配置
// 2. FeatureConfig 结构体：功能配置
// 3. BotConfig 结构体：机器人配置
// 4. FeishuConfig 结构体：飞书配置
// 5. MemosConfig 结构体：Memos 配置
// 6. Load 函数：加载配置文件
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 应用配置
// 包含应用程序的所有配置信息
//
// 此结构体是配置的顶层容器，包含日志配置、功能配置和机器人配置。
// 配置可以从 YAML 文件或环境变量加载。
type Config struct {
	Log      LogConfig       `mapstructure:"log"`      // 日志配置
	Features []FeatureConfig `mapstructure:"features"` // 功能配置列表
	Bots     []BotConfig     `mapstructure:"bots"`     // 机器人配置列表
}

// FeatureConfig 功能配置
// 定义单个功能的配置信息
//
// 每个功能都有自己的配置，包括 ID、名称、是否启用和自定义配置。
// 自定义配置通过 map[string]interface{} 传递，便于扩展。
type FeatureConfig struct {
	ID          string                 `mapstructure:"id"`          // 功能唯一标识符（用户在配置文件中使用）
	InternalID  string                 `mapstructure:"internal_id"` // 功能内部ID（代码中写死的）
	Name        string                 `mapstructure:"name"`        // 功能名称（可选，使用功能默认值）
	Description string                 `mapstructure:"description"` // 功能描述（可选，使用功能默认值）
	Enabled     bool                   `mapstructure:"enabled"`     // 是否启用（可选，默认为 true）
	Config      map[string]interface{} `mapstructure:"config"`      // 自定义配置（可选，使用功能默认值）
	// 所有功能配置都通过 Config map 传递，便于扩展
}

// BotConfig 机器人配置
// 定义单个机器人的配置信息
//
// 每个机器人都有自己的配置，包括 ID、名称、是否启用、功能映射和飞书配置。
// 支持多个机器人实例，每个机器人可以有不同的飞书应用。
type BotConfig struct {
	ID                    string                 `mapstructure:"id"`                      // 机器人唯一标识符
	Name                  string                 `mapstructure:"name"`                    // 机器人名称
	Description           string                 `mapstructure:"description"`             // 机器人描述（可选）
	Enabled               bool                   `mapstructure:"enabled"`                 // 是否启用
	WelcomeMessageEnabled bool                   `mapstructure:"welcome_message_enabled"` // 是否启用欢迎消息
	Feishu                FeishuConfig           `mapstructure:"feishu"`                  // 飞书配置
	Features              []FeatureMapping       `mapstructure:"features"`                // 功能映射列表
	Config                map[string]interface{} `mapstructure:"config"`                  // 自定义配置
}

// FeatureMapping 功能映射配置
// 定义机器人与功能的映射关系
//
// 此结构体用于指定机器人启用了哪些功能，以及哪个功能是默认功能。
type FeatureMapping struct {
	FeatureID string `mapstructure:"feature_id"` // 功能 ID
	Default   bool   `mapstructure:"default"`    // 是否为默认功能
}

// ValidateFeishuOnly 只验证飞书配置（已废弃，现在飞书配置在每个机器人中）
//
// 此方法用于验证所有启用的机器人的飞书配置是否完整。
// 如果配置不完整，返回错误。
//
// 返回值：
// - error：验证过程中的错误，成功则返回 nil
func (c *Config) ValidateFeishuOnly() error {
	// 检查所有启用的机器人的飞书配置
	for _, bot := range c.Bots {
		if bot.Enabled {
			if bot.Feishu.AppID == "" {
				return fmt.Errorf("feishu.app_id is required for bot: %s", bot.Name)
			}
			if bot.Feishu.AppSecret == "" {
				return fmt.Errorf("feishu.app_secret is required for bot: %s", bot.Name)
			}
		}
	}
	return nil
}

// FeishuConfig 飞书配置
// 定义飞书应用的配置信息
//
// 飞书配置包括应用 ID、应用密钥、验证令牌、加密密钥和是否使用 WebSocket。
// 每个机器人都可以有自己的飞书应用配置。
type FeishuConfig struct {
	AppID             string `mapstructure:"app_id"`             // 飞书应用 ID
	AppSecret         string `mapstructure:"app_secret"`         // 飞书应用密钥
	VerificationToken string `mapstructure:"verification_token"` // 验证令牌（可选）
	EncryptKey        string `mapstructure:"encrypt_key"`        // 加密密钥（可选）
	UseWebSocket      bool   `mapstructure:"use_websocket"`      // 是否使用 WebSocket（默认为 true）
}

// MemosConfig Memos 配置
// 定义 Memos 笔记系统的配置信息
//
// Memos 配置包括服务地址、访问令牌和默认可见性。
// 用于 Memos 功能连接到 Memos 服务器。
type MemosConfig struct {
	BaseURL           string `mapstructure:"base_url"`           // Memos 服务地址
	AccessToken       string `mapstructure:"access_token"`       // 访问令牌
	DefaultVisibility string `mapstructure:"default_visibility"` // 默认可见性（PUBLIC 或 PRIVATE）
}

// LogConfig 日志配置
// 定义日志的配置信息
//
// 日志配置包括日志级别和日志格式。
type LogConfig struct {
	Level  string `mapstructure:"level"`  // 日志级别（debug、info、warn、error）
	Format string `mapstructure:"format"` // 日志格式（json 或 text）
}

// Load 加载配置
//
// 此函数从配置文件或环境变量加载应用程序配置。
// 主要完成以下工作：
// 1. 设置默认值
// 2. 从环境变量读取配置
// 3. 从配置文件读取配置
// 4. 解析配置到结构体
// 5. 验证配置的完整性
//
// 参数：
// - configPath：配置文件路径，如果为空则从默认路径查找
//
// 返回值：
// - *Config：加载的配置实例
// - error：加载过程中的错误，成功则返回 nil
func Load(configPath string) (*Config, error) {
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

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 验证配置
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// setDefaults 设置默认值
//
// 此函数设置配置的默认值，当配置文件或环境变量中没有指定时使用。
//
// 参数：
// - v：Viper 实例
func setDefaults(v *viper.Viper) {
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("feishu.use_websocket", true)
	v.SetDefault("memos.default_visibility", "PRIVATE")
	v.SetDefault("features.*.enabled", true)
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
func validateConfig(cfg *Config) error {
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

// LoadFromEnv 仅从环境变量加载配置
//
// 此函数仅从环境变量加载配置，不使用配置文件。
// 适用于容器化部署环境。
//
// 返回值：
// - *Config：加载的配置实例
// - error：加载过程中的错误，成功则返回 nil
func LoadFromEnv() (*Config, error) {
	return Load("")
}
