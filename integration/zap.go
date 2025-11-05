package integration

import (
	"runtime"
	"strconv"

	elk "github.com/moonlitxy/elk_logger/pkg"
	"go.uber.org/zap/zapcore"
)

// ZapCore 实现zapcore.Core接口，将日志发送到ELK
type ZapCore struct {
	zapcore.LevelEnabler
	client *elk.Client
	fields []zapcore.Field
}

// NewZapCore 创建新的ZapCore
func NewZapCore(client *elk.Client, enabler zapcore.LevelEnabler) *ZapCore {
	return &ZapCore{
		LevelEnabler: enabler,
		client:       client,
		fields:       []zapcore.Field{},
	}
}

// With 添加字段
func (c *ZapCore) With(fields []zapcore.Field) zapcore.Core {
	clone := c.clone()
	clone.fields = append(clone.fields, fields...)
	return clone
}

// Check 检查是否应该记录日志
func (c *ZapCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return ce.AddCore(entry, c)
	}
	return ce
}

// Write 写入日志
func (c *ZapCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// 合并字段
	allFields := make([]zapcore.Field, 0, len(c.fields)+len(fields))
	allFields = append(allFields, c.fields...)
	allFields = append(allFields, fields...)

	// 转换为ELK字段
	elkFields := c.fieldsToMap(allFields)

	// 转换日志级别
	level := c.zapLevelToElkLevel(entry.Level)

	// 创建日志条目
	logEntry := elk.NewLogEntry(level, entry.Message, elkFields)
	logEntry.Logger = entry.LoggerName
	logEntry.Caller = entry.Caller.String()

	// 如果有堆栈信息
	if entry.Stack != "" {
		logEntry.Stack = entry.Stack
	}

	// 发送到ELK客户端
	return c.client.Log(level, entry.Message, elkFields)
}

// Sync 同步日志
func (c *ZapCore) Sync() error {
	c.client.Flush()
	return nil
}

// clone 克隆core
func (c *ZapCore) clone() *ZapCore {
	return &ZapCore{
		LevelEnabler: c.LevelEnabler,
		client:       c.client,
		fields:       c.fields,
	}
}

// fieldsToMap 将zap字段转换为map
func (c *ZapCore) fieldsToMap(fields []zapcore.Field) elk.Fields {
	result := make(elk.Fields)

	enc := zapcore.NewMapObjectEncoder()
	for _, field := range fields {
		field.AddTo(enc)
	}

	for k, v := range enc.Fields {
		result[k] = v
	}

	return result
}

// zapLevelToElkLevel 转换日志级别
func (c *ZapCore) zapLevelToElkLevel(level zapcore.Level) elk.LogLevel {
	switch level {
	case zapcore.DebugLevel:
		return elk.LevelDebug
	case zapcore.InfoLevel:
		return elk.LevelInfo
	case zapcore.WarnLevel:
		return elk.LevelWarn
	case zapcore.ErrorLevel:
		return elk.LevelError
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return elk.LevelFatal
	default:
		return elk.LevelInfo
	}
}

// ZapHook zap钩子实现（另一种集成方式）
type ZapHook struct {
	client *elk.Client
}

// NewZapHook 创建新的zap钩子
func NewZapHook(client *elk.Client) *ZapHook {
	return &ZapHook{
		client: client,
	}
}

// OnWrite 当写入日志时调用
func (h *ZapHook) OnWrite(entry *zapcore.Entry, fields []zapcore.Field) {
	// 转换字段
	elkFields := make(elk.Fields)
	enc := zapcore.NewMapObjectEncoder()
	for _, field := range fields {
		field.AddTo(enc)
	}
	for k, v := range enc.Fields {
		elkFields[k] = v
	}

	// 转换级别
	var level elk.LogLevel
	switch entry.Level {
	case zapcore.DebugLevel:
		level = elk.LevelDebug
	case zapcore.InfoLevel:
		level = elk.LevelInfo
	case zapcore.WarnLevel:
		level = elk.LevelWarn
	case zapcore.ErrorLevel:
		level = elk.LevelError
	default:
		level = elk.LevelFatal
	}

	// 发送日志
	_ = h.client.Log(level, entry.Message, elkFields)
}

// Helper函数：获取调用者信息
func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	return file + ":" + strconv.Itoa(line)
}
