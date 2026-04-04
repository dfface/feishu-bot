package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Feishu FeishuConfig `mapstructure:"feishu"`
	Memos  MemosConfig  `mapstructure:"memos"`
	Server ServerConfig `mapstructure:"server"`
	Log    LogConfig    `mapstructure:"log"`
}

// ValidateFeishuOnly 只验证飞书配置
func (c *Config) ValidateFeishuOnly() error {
	if c.Feishu.AppID == "" {
		return fmt.Errorf("feishu.app_id is required")
	}
	if c.Feishu.AppSecret == "" {
		return fmt.Errorf("feishu.app_secret is required")
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
	if cfg.Feishu.AppID == "" {
		return fmt.Errorf("feishu.app_id is required")
	}
	if cfg.Feishu.AppSecret == "" {
		return fmt.Errorf("feishu.app_secret is required")
	}
	if cfg.Memos.BaseURL == "" {
		return fmt.Errorf("memos.base_url is required")
	}
	if cfg.Memos.AccessToken == "" {
		return fmt.Errorf("memos.access_token is required")
	}
	return nil
}

// LoadFromEnv 仅从环境变量加载配置
func LoadFromEnv() (*Config, error) {
	return Load("")
}
