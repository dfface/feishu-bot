// utils 工具包
//
// 此包提供各种通用工具函数，用于简化代码实现。
package utils

// GetStringValue 从 map 中获取字符串值，如果不存在或类型不正确则返回默认值
//
// 参数：
// - m：包含值的 map
// - key：要获取的键
// - defaultValue：默认值
//
// 返回值：
// - string：获取到的值或默认值
func GetStringValue(m map[string]interface{}, key, defaultValue string) string {
	if value, ok := m[key].(string); ok {
		return value
	}
	return defaultValue
}
