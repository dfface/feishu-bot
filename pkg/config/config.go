package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig    `mapstructure:"server"`
	Log      LogConfig       `mapstructure:"log"`
	Features []FeatureConfig `mapstructure:"features"`
	Bots     []BotConfig     `mapstructure:"bots"`
}

// FeatureConfig 功能配置
type FeatureConfig struct {
	ID          string                 `mapstructure:"id"`
	Name        string                 `mapstructure:"name"`
	Enabled     bool                   `mapstructure:"enabled"`
	Description string                 `mapstructure:"description"`
	Config      map[string]interface{} `mapstructure:"config"`
	// 所有功能配置都通过 Config map 传递，便于扩展
}

// BotConfig 机器人配置
type BotConfig struct {
	ID      string            `mapstructure:"id"`
	Name    string            `mapstructure:"name"`
	Enabled bool              `mapstructure:"enabled"`
	Features []FeatureMapping `mapstructure:"features"`
	Config  map[string]interface{} `mapstructure:"config"`
	Feishu  FeishuConfig      `mapstructure:"feishu"` // 飞书配置
}

// FeatureMapping 功能映射配置
type FeatureMapping struct {
	FeatureID string `mapstructure:"feature_id"`
	Default   bool   `mapstructure:"default"`
}

// ValidateFeishuOnly 只验证飞书配置（已废弃，现在飞书配置在每个机器人中）
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
type FeishuConfig struct {
	AppID              string `mapstructure:"app_id"`
	AppSecret          string `mapstructure:"app_secret"`
	VerificationToken  string `mapstructure:"verification_token"`
	EncryptKey         string `mapstructure:"encrypt_key"`
	UseWebSocket       bool   `mapstructure:"use_websocket"`
}

// MemosConfig Memos 配置
type MemosConfig struct {
	BaseURL       string `mapstructure:"base_url"`
	AccessToken   string `mapstructure:"access_token"`
	DefaultVisibility string `mapstructure:"default_visibility"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int `mapstructure:"port"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// Load 加载配置
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
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 8080)
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("feishu.use_websocket", true)
	v.SetDefault("memos.default_visibility", "PRIVATE")
}

// validateConfig 验证配置
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
func LoadFromEnv() (*Config, error) {
	return Load("")
}
