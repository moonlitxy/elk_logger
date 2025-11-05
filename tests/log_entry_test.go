package tests

import (
	"encoding/json"
	"testing"
	"time"

	elk "github.com/moonlitxy/elk_logger/pkg"
)

func TestNewLogEntry(t *testing.T) {
	fields := elk.Fields{
		"user_id": 123,
		"action":  "login",
	}

	entry := elk.NewLogEntry(elk.LevelInfo, "test message", fields)

	if entry == nil {
		t.Fatal("NewLogEntry returned nil")
	}

	if entry.Level != elk.LevelInfo {
		t.Errorf("Level = %s, want %s", entry.Level, elk.LevelInfo)
	}

	if entry.Message != "test message" {
		t.Errorf("Message = %s, want test message", entry.Message)
	}

	if entry.Fields["user_id"] != 123 {
		t.Errorf("Fields[user_id] = %v, want 123", entry.Fields["user_id"])
	}
}

func TestLogEntryToJSON(t *testing.T) {
	fields := elk.Fields{
		"user_id": 123,
		"ip":      "192.168.1.1",
	}

	entry := elk.NewLogEntry(elk.LevelInfo, "test message", fields)
	entry.ServiceName = "test-service"
	entry.Environment = "test"
	entry.Logger = "test-logger"

	jsonData, err := entry.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// 解析JSON
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// 验证字段
	if result["level"] != string(elk.LevelInfo) {
		t.Errorf("level = %v, want %s", result["level"], elk.LevelInfo)
	}

	if result["message"] != "test message" {
		t.Errorf("message = %v, want test message", result["message"])
	}

	if result["service.name"] != "test-service" {
		t.Errorf("service.name = %v, want test-service", result["service.name"])
	}

	if result["user_id"].(float64) != 123 {
		t.Errorf("user_id = %v, want 123", result["user_id"])
	}
}

func TestLogEntryClone(t *testing.T) {
	fields := elk.Fields{
		"key1": "value1",
		"key2": 42,
	}

	original := elk.NewLogEntry(elk.LevelInfo, "original message", fields)
	original.ServiceName = "original-service"

	// 克隆
	cloned := original.Clone()

	// 修改克隆
	cloned.Message = "cloned message"
	cloned.ServiceName = "cloned-service"
	cloned.Fields["key1"] = "modified"

	// 验证原始对象未被修改
	if original.Message == "cloned message" {
		t.Error("original message was modified")
	}

	if original.ServiceName == "cloned-service" {
		t.Error("original service name was modified")
	}

	if original.Fields["key1"] == "modified" {
		t.Error("original fields were modified")
	}
}

func TestLogLevels(t *testing.T) {
	levels := []elk.LogLevel{
		elk.LevelDebug,
		elk.LevelInfo,
		elk.LevelWarn,
		elk.LevelError,
		elk.LevelFatal,
	}

	for _, level := range levels {
		entry := elk.NewLogEntry(level, "test", nil)
		if entry.Level != level {
			t.Errorf("Level = %s, want %s", entry.Level, level)
		}
	}
}

func TestLogEntryTimestamp(t *testing.T) {
	before := time.Now()
	entry := elk.NewLogEntry(elk.LevelInfo, "test", nil)
	after := time.Now()

	if entry.Timestamp.Before(before) || entry.Timestamp.After(after) {
		t.Error("Timestamp is not within expected range")
	}
}

func TestLogEntryWithNilFields(t *testing.T) {
	entry := elk.NewLogEntry(elk.LevelInfo, "test", nil)

	jsonData, err := entry.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON with nil fields failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result["message"] != "test" {
		t.Error("message field missing or incorrect")
	}
}
