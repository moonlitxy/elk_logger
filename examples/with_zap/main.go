package main

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/moonlitxy/elk_logger/integration"
	elk "github.com/moonlitxy/elk_logger/pkg"
)

func main() {
	// 1. 创建ELK客户端
	elkConfig := elk.DefaultConfig()
	elkConfig.ESAddresses = []string{"http://localhost:9200"}
	elkConfig.ServiceName = "zap-integration-demo"
	elkConfig.Environment = "development"

	elkClient, err := elk.NewClient(elkConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to create ELK client: %v", err))
	}
	defer elkClient.Close()

	// 2. 创建Zap配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "@timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 3. 创建多个Core
	// Console输出
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleCore := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)

	// ELK输出
	elkCore := integration.NewZapCore(elkClient, zapcore.InfoLevel)

	// 4. 组合多个Core
	core := zapcore.NewTee(
		consoleCore, // 输出到控制台
		elkCore,     // 同时输出到ELK
	)

	// 5. 创建Logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	defer logger.Sync()

	// 6. 使用Logger记录日志
	fmt.Println("Zap集成示例启动...")

	logger.Info("应用程序启动",
		zap.String("version", "2.0.0"),
		zap.Int("port", 8080),
		zap.String("mode", "production"),
	)

	logger.Debug("调试信息",
		zap.String("module", "main"),
		zap.String("function", "main"),
	)

	logger.Warn("警告信息",
		zap.String("component", "database"),
		zap.Float64("connection_pool_usage", 0.85),
	)

	// 使用结构化日志
	logger.Info("用户操作",
		zap.Int64("user_id", 10001),
		zap.String("action", "create_order"),
		zap.String("order_id", "ORD-2024-001"),
		zap.Float64("amount", 299.99),
	)

	// 嵌套字段
	logger.Info("API请求",
		zap.String("method", "POST"),
		zap.String("path", "/api/v1/users"),
		zap.Int("status_code", 200),
		zap.Duration("duration", 125000000), // 125ms
		zap.Object("request", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
			enc.AddString("client_ip", "192.168.1.100")
			enc.AddString("user_agent", "Mozilla/5.0")
			return nil
		})),
	)

	// 模拟错误日志（带堆栈）
	logger.Error("处理失败",
		zap.String("operation", "process_payment"),
		zap.Error(fmt.Errorf("payment gateway timeout")),
		zap.String("payment_id", "PAY-123456"),
	)

	// 批量日志
	fmt.Println("\n发送批量日志...")
	for i := 0; i < 20; i++ {
		logger.Info("批处理任务",
			zap.Int("task_id", i),
			zap.String("status", "completed"),
		)
	}

	fmt.Println("\n等待日志发送...")
	elkClient.Flush()

	// 打印统计信息
	metrics := elkClient.GetMetrics()
	fmt.Printf("\n--- ELK日志统计 ---\n")
	fmt.Printf("总日志数: %d\n", metrics.TotalLogs)
	fmt.Printf("成功发送: %d\n", metrics.SuccessLogs)
	fmt.Printf("发送失败: %d\n", metrics.FailedLogs)
	fmt.Printf("平均延迟: %d ms\n", metrics.AvgLatency)

	fmt.Println("\n程序执行完成！")
}
