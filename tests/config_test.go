package tests

import (
	"testing"
	"time"

	elk "github.com/moonlitxy/elk_logger/pkg"
)

func TestDefaultConfig(t *testing.T) {
	config := elk.DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}

	// 验证默认值
	if len(config.ESAddresses) == 0 {
		t.Error("ESAddresses should not be empty")
	}

	if config.BatchSize <= 0 {
		t.Error("BatchSize should be greater than 0")
	}

	if config.QueueSize <= 0 {
		t.Error("QueueSize should be greater than 0")
	}

	if config.WorkerCount <= 0 {
		t.Error("WorkerCount should be greater than 0")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *elk.Config
		expectErr bool
	}{
		{
			name:      "valid config",
			config:    elk.DefaultConfig(),
			expectErr: false,
		},
		{
			name: "empty es addresses",
			config: &elk.Config{
				ESAddresses: []string{},
				BatchSize:   100,
				QueueSize:   1000,
				WorkerCount: 4,
			},
			expectErr: true,
		},
		{
			name: "invalid batch size",
			config: &elk.Config{
				ESAddresses: []string{"http://localhost:9200"},
				BatchSize:   0,
				QueueSize:   1000,
				WorkerCount: 4,
			},
			expectErr: true,
		},
		{
			name: "invalid queue size",
			config: &elk.Config{
				ESAddresses: []string{"http://localhost:9200"},
				BatchSize:   100,
				QueueSize:   0,
				WorkerCount: 4,
			},
			expectErr: true,
		},
		{
			name: "invalid worker count",
			config: &elk.Config{
				ESAddresses: []string{"http://localhost:9200"},
				BatchSize:   100,
				QueueSize:   1000,
				WorkerCount: 0,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

func TestConfigCustomization(t *testing.T) {
	config := elk.DefaultConfig()

	// 自定义配置
	config.ESAddresses = []string{"http://es1:9200", "http://es2:9200"}
	config.ServiceName = "test-service"
	config.Environment = "test"
	config.BatchSize = 500
	config.BatchTimeout = 10 * time.Second

	if err := config.Validate(); err != nil {
		t.Errorf("customized config should be valid: %v", err)
	}

	if config.ServiceName != "test-service" {
		t.Errorf("ServiceName = %s, want test-service", config.ServiceName)
	}

	if config.Environment != "test" {
		t.Errorf("Environment = %s, want test", config.Environment)
	}
}
