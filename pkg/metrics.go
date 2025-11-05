package elk_logger

import (
	"sync/atomic"
	"time"
)

// Metrics 监控指标
type Metrics struct {
	TotalLogs   int64 // 总日志数
	SuccessLogs int64 // 成功发送数
	FailedLogs  int64 // 失败数
	DroppedLogs int64 // 丢弃数

	totalLatency int64 // 总延迟（纳秒）
	latencyCount int64 // 延迟计数
}

// NewMetrics 创建新的指标收集器
func NewMetrics() *Metrics {
	return &Metrics{}
}

// IncTotal 增加总日志数
func (m *Metrics) IncTotal() {
	atomic.AddInt64(&m.TotalLogs, 1)
}

// IncSuccess 增加成功数
func (m *Metrics) IncSuccess() {
	atomic.AddInt64(&m.SuccessLogs, 1)
}

// IncFailed 增加失败数
func (m *Metrics) IncFailed() {
	atomic.AddInt64(&m.FailedLogs, 1)
}

// IncDropped 增加丢弃数
func (m *Metrics) IncDropped() {
	atomic.AddInt64(&m.DroppedLogs, 1)
}

// RecordLatency 记录延迟
func (m *Metrics) RecordLatency(latency time.Duration) {
	atomic.AddInt64(&m.totalLatency, int64(latency))
	atomic.AddInt64(&m.latencyCount, 1)
}

// GetAvgLatency 获取平均延迟（毫秒）
func (m *Metrics) GetAvgLatency() int64 {
	count := atomic.LoadInt64(&m.latencyCount)
	if count == 0 {
		return 0
	}
	total := atomic.LoadInt64(&m.totalLatency)
	return (total / count) / int64(time.Millisecond)
}

// Snapshot 获取指标快照
func (m *Metrics) Snapshot() MetricsSnapshot {
	return MetricsSnapshot{
		TotalLogs:   atomic.LoadInt64(&m.TotalLogs),
		SuccessLogs: atomic.LoadInt64(&m.SuccessLogs),
		FailedLogs:  atomic.LoadInt64(&m.FailedLogs),
		DroppedLogs: atomic.LoadInt64(&m.DroppedLogs),
		AvgLatency:  m.GetAvgLatency(),
	}
}

// MetricsSnapshot 指标快照
type MetricsSnapshot struct {
	TotalLogs   int64 `json:"total_logs"`
	SuccessLogs int64 `json:"success_logs"`
	FailedLogs  int64 `json:"failed_logs"`
	DroppedLogs int64 `json:"dropped_logs"`
	AvgLatency  int64 `json:"avg_latency_ms"`
}

// Reset 重置指标
func (m *Metrics) Reset() {
	atomic.StoreInt64(&m.TotalLogs, 0)
	atomic.StoreInt64(&m.SuccessLogs, 0)
	atomic.StoreInt64(&m.FailedLogs, 0)
	atomic.StoreInt64(&m.DroppedLogs, 0)
	atomic.StoreInt64(&m.totalLatency, 0)
	atomic.StoreInt64(&m.latencyCount, 0)
}
