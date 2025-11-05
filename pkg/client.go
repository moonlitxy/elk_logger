package elk_logger

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

// Client ELK日志客户端
type Client struct {
	config  *Config
	sender  *Sender
	batch   *Batch
	queue   chan *LogEntry
	metrics *Metrics

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	hostName string
	hostIP   string

	closed bool
	mu     sync.Mutex
}

// NewClient 创建新的ELK客户端
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// 创建发送器
	sender, err := NewSender(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create sender: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		config:  config,
		sender:  sender,
		batch:   NewBatch(config.BatchSize, config.BatchTimeout),
		queue:   make(chan *LogEntry, config.QueueSize),
		metrics: NewMetrics(),
		ctx:     ctx,
		cancel:  cancel,
	}

	// 获取主机信息
	if config.EnableHostInfo {
		client.hostName, _ = os.Hostname()
		client.hostIP = getLocalIP()
	}

	// 启动工作协程
	client.startWorkers()

	// 启动定时刷新协程
	client.startFlusher()

	return client, nil
}

// Log 记录日志
func (c *Client) Log(level LogLevel, message string, fields Fields) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return fmt.Errorf("client is closed")
	}
	c.mu.Unlock()

	entry := NewLogEntry(level, message, fields)

	// 添加元数据
	entry.ServiceName = c.config.ServiceName
	entry.Environment = c.config.Environment

	if c.config.EnableHostInfo {
		entry.HostName = c.hostName
		entry.IP = c.hostIP
	}

	c.metrics.IncTotal()

	// 发送到队列
	select {
	case c.queue <- entry:
		return nil
	default:
		// 队列满
		if c.config.DiscardOnFull {
			c.metrics.IncDropped()
			return nil
		}

		// 阻塞等待
		select {
		case c.queue <- entry:
			return nil
		case <-time.After(5 * time.Second):
			c.metrics.IncDropped()
			return fmt.Errorf("queue is full, log dropped")
		}
	}
}

// Debug 记录Debug级别日志
func (c *Client) Debug(message string, fields Fields) error {
	return c.Log(LevelDebug, message, fields)
}

// Info 记录Info级别日志
func (c *Client) Info(message string, fields Fields) error {
	return c.Log(LevelInfo, message, fields)
}

// Warn 记录Warn级别日志
func (c *Client) Warn(message string, fields Fields) error {
	return c.Log(LevelWarn, message, fields)
}

// Error 记录Error级别日志
func (c *Client) Error(message string, fields Fields) error {
	return c.Log(LevelError, message, fields)
}

// Fatal 记录Fatal级别日志
func (c *Client) Fatal(message string, fields Fields) error {
	return c.Log(LevelFatal, message, fields)
}

// startWorkers 启动工作协程
func (c *Client) startWorkers() {
	for i := 0; i < c.config.WorkerCount; i++ {
		c.wg.Add(1)
		go c.worker()
	}
}

// worker 工作协程，从队列接收日志并添加到批次
func (c *Client) worker() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case entry := <-c.queue:
			startTime := time.Now()

			// 添加到批次
			shouldFlush := c.batch.Add(entry)

			if shouldFlush {
				c.flush()
			}

			// 记录延迟
			c.metrics.RecordLatency(time.Since(startTime))
		}
	}
}

// startFlusher 启动定时刷新协程
func (c *Client) startFlusher() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		ticker := time.NewTicker(c.config.FlushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-c.ctx.Done():
				return
			case <-ticker.C:
				if c.batch.ShouldFlush() {
					c.flush()
				}
			}
		}
	}()
}

// flush 刷新批次
func (c *Client) flush() {
	entries := c.batch.Flush()
	if len(entries) == 0 {
		return
	}

	// 发送到ES
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := c.sender.SendWithRetry(ctx, entries)
	if err != nil {
		c.metrics.IncFailed()
		// TODO: 可以将失败的日志写入本地文件
		fmt.Printf("Failed to send logs to Elasticsearch: %v\n", err)
	} else {
		c.metrics.IncSuccess()
	}
}

// Flush 手动刷新所有缓存的日志
func (c *Client) Flush() {
	c.flush()
}

// GetMetrics 获取监控指标
func (c *Client) GetMetrics() MetricsSnapshot {
	return c.metrics.Snapshot()
}

// Close 关闭客户端
func (c *Client) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	c.mu.Unlock()

	// 取消上下文
	c.cancel()

	// 等待所有工作协程退出
	c.wg.Wait()

	// 最后刷新一次
	c.flush()

	// 关闭发送器
	if c.sender != nil {
		return c.sender.Close()
	}

	return nil
}

// getLocalIP 获取本地IP地址
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}

	return ""
}
