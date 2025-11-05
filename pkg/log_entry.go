package elk_logger

import (
	"encoding/json"
	"time"
)

// LogLevel 日志级别
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
	LevelFatal LogLevel = "fatal"
)

// Fields 自定义字段类型
type Fields map[string]interface{}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp   time.Time `json:"@timestamp"`          // 日志时间戳
	Level       LogLevel  `json:"level"`               // 日志级别
	Message     string    `json:"message"`             // 日志消息
	Logger      string    `json:"logger,omitempty"`    // 日志器名称
	Caller      string    `json:"caller,omitempty"`    // 调用位置
	Stack       string    `json:"stack,omitempty"`     // 堆栈信息（错误时）
	Fields      Fields    `json:"fields,omitempty"`    // 自定义字段
	ServiceName string    `json:"service.name"`        // 服务名称
	Environment string    `json:"environment"`         // 环境（dev/test/prod）
	HostName    string    `json:"host.name,omitempty"` // 主机名
	IP          string    `json:"host.ip,omitempty"`   // IP地址
}

// NewLogEntry 创建新的日志条目
func NewLogEntry(level LogLevel, message string, fields Fields) *LogEntry {
	return &LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Fields:    fields,
	}
}

// ToJSON 转换为JSON
func (l *LogEntry) ToJSON() ([]byte, error) {
	// 将自定义字段平铺到顶层
	data := make(map[string]interface{})

	data["@timestamp"] = l.Timestamp.Format(time.RFC3339Nano)
	data["level"] = l.Level
	data["message"] = l.Message

	if l.Logger != "" {
		data["logger"] = l.Logger
	}
	if l.Caller != "" {
		data["caller"] = l.Caller
	}
	if l.Stack != "" {
		data["stack"] = l.Stack
	}
	if l.ServiceName != "" {
		data["service.name"] = l.ServiceName
	}
	if l.Environment != "" {
		data["environment"] = l.Environment
	}
	if l.HostName != "" {
		data["host.name"] = l.HostName
	}
	if l.IP != "" {
		data["host.ip"] = l.IP
	}

	// 合并自定义字段
	for k, v := range l.Fields {
		data[k] = v
	}

	return json.Marshal(data)
}

// Clone 克隆日志条目
func (l *LogEntry) Clone() *LogEntry {
	clone := *l
	if l.Fields != nil {
		clone.Fields = make(Fields, len(l.Fields))
		for k, v := range l.Fields {
			clone.Fields[k] = v
		}
	}
	return &clone
}
