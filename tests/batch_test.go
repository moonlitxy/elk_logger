package tests

import (
	"testing"
	"time"

	elk "github.com/moonlitxy/elk_logger/pkg"
)

func TestBatchAdd(t *testing.T) {
	batch := elk.NewBatch(10, 5*time.Second)

	// 添加日志
	entry := elk.NewLogEntry(elk.LevelInfo, "test message", nil)
	shouldFlush := batch.Add(entry)

	if shouldFlush {
		t.Error("should not flush after adding 1 entry when max is 10")
	}

	if batch.Size() != 1 {
		t.Errorf("batch size = %d, want 1", batch.Size())
	}
}

func TestBatchFlushOnSize(t *testing.T) {
	maxSize := 5
	batch := elk.NewBatch(maxSize, 10*time.Second)

	// 添加到达上限
	var shouldFlush bool
	for i := 0; i < maxSize; i++ {
		entry := elk.NewLogEntry(elk.LevelInfo, "test", nil)
		shouldFlush = batch.Add(entry)
	}

	if !shouldFlush {
		t.Error("should flush when batch is full")
	}

	if batch.Size() != maxSize {
		t.Errorf("batch size = %d, want %d", batch.Size(), maxSize)
	}
}

func TestBatchFlush(t *testing.T) {
	batch := elk.NewBatch(10, 5*time.Second)

	// 添加一些日志
	for i := 0; i < 5; i++ {
		entry := elk.NewLogEntry(elk.LevelInfo, "test", nil)
		batch.Add(entry)
	}

	// 刷新
	entries := batch.Flush()

	if len(entries) != 5 {
		t.Errorf("flushed entries = %d, want 5", len(entries))
	}

	if batch.Size() != 0 {
		t.Errorf("batch size after flush = %d, want 0", batch.Size())
	}
}

func TestBatchShouldFlushTimeout(t *testing.T) {
	timeout := 100 * time.Millisecond
	batch := elk.NewBatch(10, timeout)

	// 添加一个日志
	entry := elk.NewLogEntry(elk.LevelInfo, "test", nil)
	batch.Add(entry)

	// 立即检查，不应该刷新
	if batch.ShouldFlush() {
		t.Error("should not flush immediately")
	}

	// 等待超时
	time.Sleep(timeout + 10*time.Millisecond)

	// 现在应该刷新
	if !batch.ShouldFlush() {
		t.Error("should flush after timeout")
	}
}

func TestBatchClear(t *testing.T) {
	batch := elk.NewBatch(10, 5*time.Second)

	// 添加日志
	for i := 0; i < 3; i++ {
		entry := elk.NewLogEntry(elk.LevelInfo, "test", nil)
		batch.Add(entry)
	}

	// 清空
	batch.Clear()

	if batch.Size() != 0 {
		t.Errorf("batch size after clear = %d, want 0", batch.Size())
	}
}

func TestBatchConcurrency(t *testing.T) {
	batch := elk.NewBatch(1000, 5*time.Second)

	// 并发添加
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				entry := elk.NewLogEntry(elk.LevelInfo, "test", nil)
				batch.Add(entry)
			}
			done <- true
		}()
	}

	// 等待完成
	for i := 0; i < 10; i++ {
		<-done
	}

	if batch.Size() != 1000 {
		t.Errorf("batch size = %d, want 1000", batch.Size())
	}
}
