package main

import (
	"fmt"
	"time"

	elk "github.com/moonlitxy/elk_logger/pkg"
)

func main() {
	// 创建配置
	config := elk.DefaultConfig()
	config.ESAddresses = []string{"http://localhost:9200"}
	config.ServiceName = "my-application"
	config.Environment = "development"
	config.IndexPattern = "logs-{date}"

	// 创建ELK客户端
	client, err := elk.NewClient(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create ELK client: %v", err))
	}
	defer client.Close()

	fmt.Println("ELK Logger示例启动...")

	// 1. 基本日志记录
	client.Info("应用程序启动", elk.Fields{
		"version": "1.0.0",
		"port":    8080,
	})

	// 2. Debug级别日志
	client.Debug("调试信息", elk.Fields{
		"module": "main",
		"action": "initialization",
	})

	// 3. 警告日志
	client.Warn("这是一个警告", elk.Fields{
		"reason": "磁盘使用率较高",
		"usage":  85.5,
	})

	// 4. 错误日志
	client.Error("处理请求失败", elk.Fields{
		"request_id": "req-123456",
		"error":      "connection timeout",
		"retry":      3,
	})

	// 5. 带复杂字段的日志
	client.Info("用户登录", elk.Fields{
		"user_id":    12345,
		"username":   "zhangsan",
		"ip":         "192.168.1.100",
		"user_agent": "Mozilla/5.0",
		"login_time": time.Now(),
		"metadata": map[string]interface{}{
			"device": "mobile",
			"os":     "iOS",
		},
	})

	// 6. 模拟批量日志
	fmt.Println("发送批量日志...")
	for i := 0; i < 50; i++ {
		client.Info(fmt.Sprintf("批量日志 #%d", i), elk.Fields{
			"index":     i,
			"timestamp": time.Now().Unix(),
		})
		time.Sleep(10 * time.Millisecond)
	}

	// 手动刷新确保所有日志都被发送
	fmt.Println("刷新日志缓冲区...")
	client.Flush()

	// 获取并打印指标
	metrics := client.GetMetrics()
	fmt.Printf("\n--- 日志统计 ---\n")
	fmt.Printf("总日志数: %d\n", metrics.TotalLogs)
	fmt.Printf("成功发送: %d\n", metrics.SuccessLogs)
	fmt.Printf("发送失败: %d\n", metrics.FailedLogs)
	fmt.Printf("丢弃数量: %d\n", metrics.DroppedLogs)
	fmt.Printf("平均延迟: %d ms\n", metrics.AvgLatency)

	fmt.Println("\n程序执行完成！")
}
