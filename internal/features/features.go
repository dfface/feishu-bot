package features

import (
	"context"

	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"

	echoFeature "github.com/dfface/feishu-bot/internal/features/echo"
	memosFeature "github.com/dfface/feishu-bot/internal/features/memos"
	"github.com/dfface/feishu-bot/internal/config"
)

// Feature 功能接口
type Feature interface {
	// ID 返回功能唯一标识符
	ID() string
	// Name 返回功能名称
	Name() string
	// Description 返回功能描述
	Description() string
	// HandleMessage 处理消息
	HandleMessage(ctx context.Context, event *larkim.P2MessageReceiveV1) error
	// MatchPrefix 返回匹配前缀
	MatchPrefix() string
	// Initialize 初始化功能
	Initialize(featureConfig *config.FeatureConfig) error
}

// FeatureRegistry 功能注册中心
type FeatureRegistry struct {
	features map[string]Feature
}

// NewFeatureRegistry 创建功能注册中心
func NewFeatureRegistry() *FeatureRegistry {
	return &FeatureRegistry{
		features: make(map[string]Feature),
	}
}

// Register 注册功能
func (r *FeatureRegistry) Register(feature Feature) {
	r.features[feature.ID()] = feature
}

// Get 根据ID获取功能
func (r *FeatureRegistry) Get(id string) Feature {
	return r.features[id]
}

// GetAll 获取所有功能
func (r *FeatureRegistry) GetAll() []Feature {
	var features []Feature
	for _, f := range r.features {
		features = append(features, f)
	}
	return features
}

// RegisterFeatures 注册所有功能
func RegisterFeatures(registry *FeatureRegistry) {
	// 注册回声功能
	registry.Register(echoFeature.NewEchoFeature())
	
	// 注册Memos功能
	registry.Register(memosFeature.NewMemosFeature())
}
