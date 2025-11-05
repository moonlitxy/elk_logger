package integration

import (
	elk "github.com/moonlitxy/elk_logger/pkg"
)

// LogrusHook logrus钩子，将日志发送到ELK
type LogrusHook struct {
	client *elk.Client
}

// NewLogrusHook 创建新的logrus钩子
func NewLogrusHook(client *elk.Client) *LogrusHook {
	return &LogrusHook{
		client: client,
	}
}

// Levels 返回支持的日志级别
func (h *LogrusHook) Levels() []interface{} {
	// 这里返回interface{}类型，避免导入logrus包
	// 实际使用时会匹配logrus.Level类型
	return []interface{}{
		"panic", "fatal", "error", "warn", "warning", "info", "debug", "trace",
	}
}

// Fire 触发钩子
func (h *LogrusHook) Fire(entry interface{}) error {
	// 注意：这里使用interface{}避免导入logrus
	// 实际使用时，调用方会传入*logrus.Entry

	// 由于不能直接导入logrus，这里提供一个通用的转换函数
	// 使用者需要自己实现LogrusEntry到LogEntry的转换

	// 这是一个示例实现框架
	// 实际使用时需要用户自己根据logrus.Entry进行类型断言和转换

	return nil
}

// LogrusEntryConverter logrus条目转换器接口
// 用户需要实现这个接口来转换logrus.Entry
type LogrusEntryConverter interface {
	Convert(entry interface{}) (*elk.LogEntry, error)
}

// DefaultLogrusConverter 默认的logrus转换器（示例）
type DefaultLogrusConverter struct{}

// Convert 转换logrus条目
// 注意：这是一个示例，实际需要导入logrus包
func (c *DefaultLogrusConverter) Convert(entry interface{}) (*elk.LogEntry, error) {
	// 这里应该进行类型断言
	// logEntry, ok := entry.(*logrus.Entry)
	// if !ok {
	//     return nil, fmt.Errorf("invalid entry type")
	// }

	// 转换逻辑...
	// return &elk.LogEntry{
	//     Timestamp: logEntry.Time,
	//     Level:     convertLevel(logEntry.Level),
	//     Message:   logEntry.Message,
	//     Fields:    elk.Fields(logEntry.Data),
	// }, nil

	return nil, nil
}

// 辅助函数：logrus级别转ELK级别
func convertLogrusLevel(level string) elk.LogLevel {
	switch level {
	case "debug", "trace":
		return elk.LevelDebug
	case "info":
		return elk.LevelInfo
	case "warn", "warning":
		return elk.LevelWarn
	case "error":
		return elk.LevelError
	case "fatal", "panic":
		return elk.LevelFatal
	default:
		return elk.LevelInfo
	}
}
