package elk_logger

import "time"

// Config ELK日志采集器配置
type Config struct {
	// Elasticsearch配置
	ESAddresses  []string `json:"es_addresses"`  // ES集群地址
	ESUsername   string   `json:"es_username"`   // ES用户名
	ESPassword   string   `json:"es_password"`   // ES密码
	IndexPattern string   `json:"index_pattern"` // 索引模式，如 "logs-{date}"

	// 批量发送配置
	BatchSize     int           `json:"batch_size"`     // 批量大小（条数）
	BatchTimeout  time.Duration `json:"batch_timeout"`  // 批量超时时间
	FlushInterval time.Duration `json:"flush_interval"` // 强制刷新间隔

	// 队列配置
	QueueSize   int `json:"queue_size"`   // 队列大小
	WorkerCount int `json:"worker_count"` // 工作协程数

	// 重试配置
	RetryCount    int           `json:"retry_count"`    // 重试次数
	RetryInterval time.Duration `json:"retry_interval"` // 重试间隔

	// 应用信息
	ServiceName    string `json:"service_name"`     // 服务名称
	Environment    string `json:"environment"`      // 环境
	EnableHostInfo bool   `json:"enable_host_info"` // 是否添加主机信息

	// 高级配置
	EnableCompression bool          `json:"enable_compression"` // 是否启用压缩
	MaxRetryBackoff   time.Duration `json:"max_retry_backoff"`  // 最大重试退避时间
	DiscardOnFull     bool          `json:"discard_on_full"`    // 队列满时是否丢弃
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		ESAddresses:       []string{"http://localhost:9200"},
		IndexPattern:      "logs-{date}",
		BatchSize:         100,
		BatchTimeout:      5 * time.Second,
		FlushInterval:     10 * time.Second,
		QueueSize:         10000,
		WorkerCount:       4,
		RetryCount:        3,
		RetryInterval:     1 * time.Second,
		MaxRetryBackoff:   30 * time.Second,
		ServiceName:       "unknown-service",
		Environment:       "development",
		EnableHostInfo:    true,
		EnableCompression: true,
		DiscardOnFull:     false,
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if len(c.ESAddresses) == 0 {
		return ErrInvalidConfig{msg: "es_addresses cannot be empty"}
	}
	if c.BatchSize <= 0 {
		return ErrInvalidConfig{msg: "batch_size must be greater than 0"}
	}
	if c.QueueSize <= 0 {
		return ErrInvalidConfig{msg: "queue_size must be greater than 0"}
	}
	if c.WorkerCount <= 0 {
		return ErrInvalidConfig{msg: "worker_count must be greater than 0"}
	}
	return nil
}

// ErrInvalidConfig 配置错误
type ErrInvalidConfig struct {
	msg string
}

func (e ErrInvalidConfig) Error() string {
	return "invalid config: " + e.msg
}
