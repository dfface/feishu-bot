package bot

import (
	"context"
	"fmt"

	"github.com/larksuite/oapi-sdk-go/v3"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	"go.uber.org/zap"
)

// Bot 定义了飞书机器人的基础接口
type Bot interface {
	// Name 返回机器人名称
	Name() string
	// GetClient 获取飞书客户端
	GetClient() *lark.Client
	// GetDispatcher 获取事件分发器
	GetDispatcher() *dispatcher.EventDispatcher
	// HandleMessage 处理消息事件
	HandleMessage(ctx context.Context, event interface{}) error
	// HandleCardAction 处理卡片交互事件
	HandleCardAction(ctx context.Context, event interface{}) error
}



// BaseBot 基础机器人实现
type BaseBot struct {
	name       string
	client     *lark.Client
	dispatcher *dispatcher.EventDispatcher
}

// NewBaseBot 创建基础机器人
func NewBaseBot(name string, client *lark.Client, verificationToken, encryptKey string) *BaseBot {
	bot := &BaseBot{
		name:   name,
		client: client,
	}

	// 设置事件分发器
	// WebSocket 模式下，verificationToken 和 encryptKey 应该为空
	bot.dispatcher = dispatcher.NewEventDispatcher("", "")

	return bot
}

// Name 返回机器人名称
func (b *BaseBot) Name() string {
	return b.name
}

// GetClient 获取飞书客户端
func (b *BaseBot) GetClient() *lark.Client {
	return b.client
}

// GetDispatcher 获取事件分发器
func (b *BaseBot) GetDispatcher() *dispatcher.EventDispatcher {
	return b.dispatcher
}

// HandleMessage 处理消息事件（默认实现）
func (b *BaseBot) HandleMessage(ctx context.Context, event interface{}) error {
	return nil
}

// HandleCardAction 处理卡片交互事件（默认实现）
func (b *BaseBot) HandleCardAction(ctx context.Context, event interface{}) error {
	return nil
}

// Manager 机器人管理器
type Manager struct {
	bots   map[string]Bot
	logger *zap.Logger
}

// NewManager 创建机器人管理器
func NewManager(logger *zap.Logger) *Manager {
	return &Manager{
		bots:   make(map[string]Bot),
		logger: logger,
	}
}

// RegisterBot 注册机器人
func (m *Manager) RegisterBot(bot Bot) {
	m.bots[bot.Name()] = bot
	m.logger.Info("Bot registered", zap.String("bot_name", bot.Name()))
}

// GetBot 获取指定名称的机器人
func (m *Manager) GetBot(name string) (Bot, bool) {
	bot, ok := m.bots[name]
	return bot, ok
}

// GetAllBots 获取所有机器人
func (m *Manager) GetAllBots() map[string]Bot {
	return m.bots
}

// GetDefaultBot 获取默认机器人
func (m *Manager) GetDefaultBot() (Bot, bool) {
	// 默认返回第一个注册的机器人
	for _, bot := range m.bots {
		return bot, true
	}
	return nil, false
}

// GetDispatcherByBotName 根据机器人名称获取事件分发器
func (m *Manager) GetDispatcherByBotName(name string) (*dispatcher.EventDispatcher, error) {
	bot, ok := m.GetBot(name)
	if !ok {
		return nil, fmt.Errorf("bot not found: %s", name)
	}
	return bot.GetDispatcher(), nil
}

// GetDefaultDispatcher 获取默认机器人的事件分发器
func (m *Manager) GetDefaultDispatcher() (*dispatcher.EventDispatcher, error) {
	bot, ok := m.GetDefaultBot()
	if !ok {
		return nil, fmt.Errorf("no bot registered")
	}
	return bot.GetDispatcher(), nil
}
