package elk_logger

import (
	"sync"
	"time"
)

// Batch 批量管理器
type Batch struct {
	mu        sync.Mutex
	entries   []*LogEntry
	maxSize   int
	timeout   time.Duration
	lastFlush time.Time
}

// NewBatch 创建新的批量管理器
func NewBatch(maxSize int, timeout time.Duration) *Batch {
	return &Batch{
		entries:   make([]*LogEntry, 0, maxSize),
		maxSize:   maxSize,
		timeout:   timeout,
		lastFlush: time.Now(),
	}
}

// Add 添加日志条目
// 返回值：是否需要刷新
func (b *Batch) Add(entry *LogEntry) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.entries = append(b.entries, entry)

	// 检查是否需要刷新
	return len(b.entries) >= b.maxSize
}

// ShouldFlush 判断是否应该刷新（基于超时）
func (b *Batch) ShouldFlush() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.entries) == 0 {
		return false
	}

	return time.Since(b.lastFlush) >= b.timeout
}

// Flush 刷新批次，返回所有日志条目
func (b *Batch) Flush() []*LogEntry {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.entries) == 0 {
		return nil
	}

	// 交换缓冲区
	entries := b.entries
	b.entries = make([]*LogEntry, 0, b.maxSize)
	b.lastFlush = time.Now()

	return entries
}

// Size 返回当前批次大小
func (b *Batch) Size() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.entries)
}

// Clear 清空批次
func (b *Batch) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.entries = make([]*LogEntry, 0, b.maxSize)
	b.lastFlush = time.Now()
}
