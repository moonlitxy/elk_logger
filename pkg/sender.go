package elk_logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

// Sender ES发送器
type Sender struct {
	client       *elasticsearch.Client
	indexPattern string
	config       *Config
}

// NewSender 创建新的发送器
func NewSender(config *Config) (*Sender, error) {
	esConfig := elasticsearch.Config{
		Addresses: config.ESAddresses,
		Username:  config.ESUsername,
		Password:  config.ESPassword,

		// 连接池配置
		MaxRetries: config.RetryCount,
		RetryBackoff: func(i int) time.Duration {
			// 指数退避
			d := time.Duration(i) * config.RetryInterval
			if d > config.MaxRetryBackoff {
				return config.MaxRetryBackoff
			}
			return d
		},

		// 启用压缩
		EnableMetrics:       true,
		EnableDebugLogger:   false,
		CompressRequestBody: config.EnableCompression,
	}

	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.Ping(client.Ping.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to ping elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch ping returned error: %s", res.Status())
	}

	return &Sender{
		client:       client,
		indexPattern: config.IndexPattern,
		config:       config,
	}, nil
}

// Send 发送日志批次到ES
func (s *Sender) Send(ctx context.Context, entries []*LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	// 构建批量请求
	var buf bytes.Buffer
	for _, entry := range entries {
		// 索引元数据
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": s.getIndexName(entry.Timestamp),
			},
		}
		metaJSON, _ := json.Marshal(meta)
		buf.Write(metaJSON)
		buf.WriteByte('\n')

		// 文档数据
		docJSON, err := entry.ToJSON()
		if err != nil {
			return fmt.Errorf("failed to marshal log entry: %w", err)
		}
		buf.Write(docJSON)
		buf.WriteByte('\n')
	}

	// 发送批量请求
	res, err := s.client.Bulk(
		bytes.NewReader(buf.Bytes()),
		s.client.Bulk.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to send bulk request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk request returned error: %s", res.Status())
	}

	// 解析响应检查是否有错误
	var bulkRes BulkResponse
	if err := json.NewDecoder(res.Body).Decode(&bulkRes); err != nil {
		return fmt.Errorf("failed to parse bulk response: %w", err)
	}

	if bulkRes.Errors {
		// 有部分失败，但不返回错误，让上层决定如何处理
		// 可以在这里记录详细的错误信息
		return fmt.Errorf("bulk request has errors")
	}

	return nil
}

// SendWithRetry 带重试的发送
func (s *Sender) SendWithRetry(ctx context.Context, entries []*LogEntry) error {
	var lastErr error

	for i := 0; i <= s.config.RetryCount; i++ {
		if i > 0 {
			// 重试前等待
			backoff := time.Duration(i) * s.config.RetryInterval
			if backoff > s.config.MaxRetryBackoff {
				backoff = s.config.MaxRetryBackoff
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		err := s.Send(ctx, entries)
		if err == nil {
			return nil
		}

		lastErr = err
	}

	return fmt.Errorf("failed after %d retries: %w", s.config.RetryCount, lastErr)
}

// getIndexName 根据时间戳生成索引名
func (s *Sender) getIndexName(timestamp time.Time) string {
	indexName := s.indexPattern
	indexName = strings.ReplaceAll(indexName, "{date}", timestamp.Format("2006.01.02"))
	indexName = strings.ReplaceAll(indexName, "{year}", timestamp.Format("2006"))
	indexName = strings.ReplaceAll(indexName, "{month}", timestamp.Format("01"))
	indexName = strings.ReplaceAll(indexName, "{day}", timestamp.Format("02"))
	return indexName
}

// Close 关闭发送器
func (s *Sender) Close() error {
	// go-elasticsearch客户端不需要显式关闭
	return nil
}

// BulkResponse ES批量响应
type BulkResponse struct {
	Errors bool                          `json:"errors"`
	Items  []map[string]BulkResponseItem `json:"items"`
}

// BulkResponseItem 批量响应项
type BulkResponseItem struct {
	Index  string `json:"_index"`
	ID     string `json:"_id"`
	Status int    `json:"status"`
	Error  *struct {
		Type   string `json:"type"`
		Reason string `json:"reason"`
	} `json:"error,omitempty"`
}
