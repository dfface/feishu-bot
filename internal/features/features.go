// features 功能管理包
//
// 此包定义了功能接口和功能注册中心，用于管理和注册各种机器人功能。
// 主要包含：
// 1. Feature 接口：定义所有功能必须实现的核心方法
// 2. FeatureRegistry 结构体：功能注册中心，用于管理和查找功能
// 3. RegisterFeatures 函数：注册所有内置功能
package features

import (
	"context"

	"github.com/dfface/feishu-bot/internal/bot"
	"github.com/dfface/feishu-bot/internal/config"
	echoFeature "github.com/dfface/feishu-bot/internal/features/echo"
	memosFeature "github.com/dfface/feishu-bot/internal/features/memos"
	"github.com/dfface/feishu-bot/internal/message"
)

// Feature 功能接口
// 定义所有功能必须实现的核心方法
//
// 此接口是所有功能实现的基础，确保所有功能都具备基本的能力。
// 任何实现了此接口的结构体都可以被视为一个完整的功能。
// 功能是机器人处理消息的基本单元，每个功能负责处理特定类型的消息。
type Feature interface {
	// ID 返回功能唯一标识符
	//
	// 返回值：
	// - string：功能的唯一标识符，用于在配置中引用和查找功能
	ID() string

	// Name 返回功能名称
	//
	// 返回值：
	// - string：功能的名称，用于显示和日志记录
	Name() string

	// Description 返回功能描述
	//
	// 返回值：
	// - string：功能的描述信息，用于说明功能的作用和用途
	Description() string

	// HandleMessage 处理消息
	//
	// 此方法是功能的核心，负责处理接收到的消息。
	// 每个功能根据自己的业务逻辑处理消息，并返回处理结果。
	//
	// 参数：
	// - ctx：上下文，用于控制请求的生命周期
	// - msgContent：消息内容，包含解析后的消息文本和其他信息
	//
	// 返回值：
	// - error：处理过程中的错误，成功则返回 nil
	HandleMessage(ctx context.Context, msgContent *message.MessageContent) error

	// MatchPrefix 返回匹配前缀
	//
	// 匹配前缀用于识别消息是否应该由此功能处理。
	// 当消息文本以此前缀开头时，该功能将被调用。
	//
	// 返回值：
	// - string：匹配前缀，例如 "!echo" 或 "!memos"
	MatchPrefix() string

	// Initialize 初始化功能
	//
	// 此方法在功能注册时被调用，用于初始化功能所需的配置和资源。
	// 每个功能可以根据自己的需求从配置中读取必要的参数。
	//
	// 参数：
	// - featureConfig：功能配置，包含功能所需的配置信息
	//
	// 返回值：
	// - error：初始化过程中的错误，成功则返回 nil
	Initialize(featureConfig *config.FeatureConfig) error

	// SetBaseBot 设置基础机器人实例
	//
	// 此方法是必需的，用于为功能提供访问机器人基础功能的能力。
	// 功能通过 BaseBot 可以访问消息处理器、消息发送器、文件上传器等组件。
	//
	// 参数：
	// - baseBot：基础机器人实例，提供各种便捷方法
	SetBaseBot(baseBot *bot.BaseBot)
}

// FeatureRegistry 功能注册中心
// 用于管理和查找功能
//
// 功能注册中心是一个中心化的功能管理器，负责存储和查找所有注册的功能。
// 它提供了注册、获取和遍历功能的方法，是功能管理的基础设施。
type FeatureRegistry struct {
	features map[string]Feature // 功能映射表，键为功能 ID，值为功能实例
}

// NewFeatureRegistry 创建功能注册中心
//
// 此函数创建一个新的功能注册中心实例，初始化内部的功能映射表。
//
// 返回值：
// - *FeatureRegistry：创建的功能注册中心实例
func NewFeatureRegistry() *FeatureRegistry {
	return &FeatureRegistry{
		features: make(map[string]Feature),
	}
}

// Register 注册功能
//
// 此方法将功能注册到功能注册中心，使其可以被查找和使用。
// 如果功能 ID 已存在，将覆盖之前的功能。
//
// 参数：
// - feature：要注册的功能实例
func (r *FeatureRegistry) Register(feature Feature) {
	r.features[feature.ID()] = feature
}

// Get 根据ID获取功能
//
// 此方法根据功能 ID 从功能注册中心查找功能。
// 如果功能不存在，返回 nil。
//
// 参数：
// - id：功能 ID
//
// 返回值：
// - Feature：找到的功能实例，如果不存在则返回 nil
func (r *FeatureRegistry) Get(id string) Feature {
	return r.features[id]
}

// GetAll 获取所有功能
//
// 此方法返回所有已注册的功能列表。
//
// 返回值：
// - []Feature：所有已注册的功能列表
func (r *FeatureRegistry) GetAll() []Feature {
	var features []Feature
	for _, f := range r.features {
		features = append(features, f)
	}
	return features
}

// RegisterFeatures 注册所有功能
//
// 此函数注册所有内置功能到功能注册中心。
// 这是功能注册的入口点，在机器人启动时被调用。
//
// 参数：
// - registry：功能注册中心实例
func RegisterFeatures(registry *FeatureRegistry) {
	// 注册回声功能
	registry.Register(echoFeature.NewEchoFeature())

	// 注册Memos功能
	registry.Register(memosFeature.NewMemosFeature())
}
