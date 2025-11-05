package tests

import (
	"testing"
	"time"

	elk "github.com/moonlitxy/elk_logger/pkg"
)

func TestMetricsIncrement(t *testing.T) {
	metrics := elk.NewMetrics()

	metrics.IncTotal()
	metrics.IncSuccess()
	metrics.IncFailed()
	metrics.IncDropped()

	snapshot := metrics.Snapshot()

	if snapshot.TotalLogs != 1 {
		t.Errorf("TotalLogs = %d, want 1", snapshot.TotalLogs)
	}

	if snapshot.SuccessLogs != 1 {
		t.Errorf("SuccessLogs = %d, want 1", snapshot.SuccessLogs)
	}

	if snapshot.FailedLogs != 1 {
		t.Errorf("FailedLogs = %d, want 1", snapshot.FailedLogs)
	}

	if snapshot.DroppedLogs != 1 {
		t.Errorf("DroppedLogs = %d, want 1", snapshot.DroppedLogs)
	}
}

func TestMetricsLatency(t *testing.T) {
	metrics := elk.NewMetrics()

	// 记录一些延迟
	metrics.RecordLatency(100 * time.Millisecond)
	metrics.RecordLatency(200 * time.Millisecond)
	metrics.RecordLatency(300 * time.Millisecond)

	avgLatency := metrics.GetAvgLatency()

	// 平均值应该是200ms
	if avgLatency != 200 {
		t.Errorf("AvgLatency = %d ms, want 200 ms", avgLatency)
	}
}

func TestMetricsReset(t *testing.T) {
	metrics := elk.NewMetrics()

	// 增加一些计数
	metrics.IncTotal()
	metrics.IncSuccess()
	metrics.RecordLatency(100 * time.Millisecond)

	// 重置
	metrics.Reset()

	snapshot := metrics.Snapshot()

	if snapshot.TotalLogs != 0 {
		t.Errorf("TotalLogs after reset = %d, want 0", snapshot.TotalLogs)
	}

	if snapshot.SuccessLogs != 0 {
		t.Errorf("SuccessLogs after reset = %d, want 0", snapshot.SuccessLogs)
	}

	if snapshot.AvgLatency != 0 {
		t.Errorf("AvgLatency after reset = %d, want 0", snapshot.AvgLatency)
	}
}

func TestMetricsConcurrency(t *testing.T) {
	metrics := elk.NewMetrics()

	// 并发增加计数
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				metrics.IncTotal()
				metrics.RecordLatency(10 * time.Millisecond)
			}
			done <- true
		}()
	}

	// 等待完成
	for i := 0; i < 10; i++ {
		<-done
	}

	snapshot := metrics.Snapshot()

	if snapshot.TotalLogs != 10000 {
		t.Errorf("TotalLogs = %d, want 10000", snapshot.TotalLogs)
	}
}

func TestMetricsSnapshot(t *testing.T) {
	metrics := elk.NewMetrics()

	metrics.IncTotal()
	metrics.IncTotal()
	metrics.IncSuccess()

	snapshot1 := metrics.Snapshot()
	snapshot2 := metrics.Snapshot()

	// 两次快照应该相同
	if snapshot1.TotalLogs != snapshot2.TotalLogs {
		t.Error("Snapshots should be identical")
	}

	// 修改指标后快照应该不同
	metrics.IncTotal()
	snapshot3 := metrics.Snapshot()

	if snapshot1.TotalLogs == snapshot3.TotalLogs {
		t.Error("Snapshot should reflect new metrics")
	}
}
