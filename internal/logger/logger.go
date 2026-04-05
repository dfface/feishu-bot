// logger 日志包
//
// 此包提供统一的日志管理功能，包括日志初始化和常用的日志方法。
// 所有模块都可以直接使用此包中的方法进行日志记录，而不需要传递 logger 对象。
package logger

import (
	"github.com/dfface/feishu-bot/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// Init 初始化日志系统
//
// 此函数初始化日志系统，设置日志级别和格式。
//
// 参数：
// - cfg：日志配置
func Init(cfg config.LogConfig) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(cfg.Level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	zapConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapLevel),
		Development:      false,
		Encoding:         cfg.Format,
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	if cfg.Format == "console" {
		zapConfig.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	var err error
	log, err = zapConfig.Build()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
}

// GetLogger 获取日志记录器
//
// 此函数返回当前的日志记录器实例。
//
// 返回值：
// - *zap.Logger：日志记录器实例
func GetLogger() *zap.Logger {
	if log == nil {
		// 如果日志未初始化，使用默认配置
		Init(config.LogConfig{
			Level:  "info",
			Format: "json",
		})
	}
	return log
}

// Debug 记录调试级别日志
//
// 参数：
// - msg：日志消息
// - fields：日志字段
func Debug(msg string, fields ...zapcore.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info 记录信息级别日志
//
// 参数：
// - msg：日志消息
// - fields：日志字段
func Info(msg string, fields ...zapcore.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn 记录警告级别日志
//
// 参数：
// - msg：日志消息
// - fields：日志字段
func Warn(msg string, fields ...zapcore.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error 记录错误级别日志
//
// 参数：
// - msg：日志消息
// - fields：日志字段
func Error(msg string, fields ...zapcore.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal 记录致命级别日志并退出程序
//
// 参数：
// - msg：日志消息
// - fields：日志字段
func Fatal(msg string, fields ...zapcore.Field) {
	GetLogger().Fatal(msg, fields...)
}

// Sync 同步日志缓冲区
//
// 此函数将日志缓冲区中的内容刷新到输出设备。
func Sync() error {
	if log != nil {
		return log.Sync()
	}
	return nil
}
