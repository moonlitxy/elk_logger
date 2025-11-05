package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	elk "github.com/moonlitxy/elk_logger/pkg"
)

// 全局ELK日志客户端
var elkLogger *elk.Client

// User 用户结构
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func main() {
	// ========== 第1步：初始化ELK日志客户端 ==========
	fmt.Println("正在初始化ELK日志客户端...")

	config := elk.DefaultConfig()
	config.ESAddresses = []string{"http://localhost:9200"}
	config.ServiceName = "user-service"
	config.Environment = "development"
	config.BatchSize = 100
	config.QueueSize = 10000

	var err error
	elkLogger, err = elk.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create ELK logger: %v", err)
	}
	// 确保程序退出时关闭客户端，发送完所有日志
	defer elkLogger.Close()

	// 记录服务启动日志
	elkLogger.Info("用户服务启动", elk.Fields{
		"version": "1.0.0",
		"port":    8080,
	})

	// ========== 第2步：注册HTTP路由 ==========
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/users", usersHandler)
	http.HandleFunc("/api/users/create", createUserHandler)
	http.HandleFunc("/health", healthHandler)

	// ========== 第3步：启动定时任务（展示后台任务日志） ==========
	go backgroundTask()

	// ========== 第4步：启动HTTP服务器 ==========
	fmt.Println("========================================")
	fmt.Println("🚀 用户服务已启动")
	fmt.Println("📡 监听端口: 8080")
	fmt.Println("📊 访问 http://localhost:8080")
	fmt.Println("🏥 健康检查 http://localhost:8080/health")
	fmt.Println("========================================")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		elkLogger.Fatal("服务启动失败", elk.Fields{
			"error": err.Error(),
		})
		log.Fatal(err)
	}
}

// homeHandler 首页处理器
func homeHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// 记录请求日志
	elkLogger.Info("收到首页请求", elk.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"client_ip":  r.RemoteAddr,
		"user_agent": r.UserAgent(),
	})

	response := map[string]interface{}{
		"service": "user-service",
		"version": "1.0.0",
		"status":  "running",
		"endpoints": []string{
			"GET  /",
			"GET  /api/users",
			"POST /api/users/create",
			"GET  /health",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	// 记录响应日志
	elkLogger.Info("首页请求完成", elk.Fields{
		"method":      r.Method,
		"path":        r.URL.Path,
		"status":      200,
		"duration_ms": time.Since(start).Milliseconds(),
	})
}

// usersHandler 获取用户列表
func usersHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	elkLogger.Info("获取用户列表", elk.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	// 模拟数据库查询
	users := []User{
		{ID: 1, Username: "zhangsan", Email: "zhangsan@example.com"},
		{ID: 2, Username: "lisi", Email: "lisi@example.com"},
		{ID: 3, Username: "wangwu", Email: "wangwu@example.com"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    users,
		"count":   len(users),
	})

	elkLogger.Info("用户列表查询成功", elk.Fields{
		"count":       len(users),
		"duration_ms": time.Since(start).Milliseconds(),
	})
}

// createUserHandler 创建用户
func createUserHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	if r.Method != http.MethodPost {
		elkLogger.Warn("错误的请求方法", elk.Fields{
			"expected": "POST",
			"actual":   r.Method,
		})
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		elkLogger.Error("解析请求数据失败", elk.Fields{
			"error": err.Error(),
		})
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 模拟创建用户
	user.ID = int(time.Now().Unix())

	elkLogger.Info("用户创建成功", elk.Fields{
		"user_id":     user.ID,
		"username":    user.Username,
		"email":       user.Email,
		"duration_ms": time.Since(start).Milliseconds(),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "用户创建成功",
		"data":    user,
	})
}

// healthHandler 健康检查
func healthHandler(w http.ResponseWriter, r *http.Request) {
	// 获取监控指标
	metrics := elkLogger.GetMetrics()

	health := map[string]interface{}{
		"status":         "healthy",
		"timestamp":      time.Now().Format(time.RFC3339),
		"uptime_seconds": time.Since(startTime).Seconds(),
		"elk_metrics": map[string]interface{}{
			"total_logs":   metrics.TotalLogs,
			"success_logs": metrics.SuccessLogs,
			"failed_logs":  metrics.FailedLogs,
			"dropped_logs": metrics.DroppedLogs,
			"avg_latency":  metrics.AvgLatency,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)

	elkLogger.Debug("健康检查", elk.Fields{
		"metrics": metrics,
	})
}

// backgroundTask 后台定时任务示例
func backgroundTask() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 模拟后台任务
		elkLogger.Info("执行后台任务", elk.Fields{
			"task_type": "cleanup",
			"timestamp": time.Now().Unix(),
		})

		// 模拟一些处理
		time.Sleep(100 * time.Millisecond)

		elkLogger.Info("后台任务完成", elk.Fields{
			"task_type": "cleanup",
			"status":    "success",
		})
	}
}

var startTime = time.Now()
