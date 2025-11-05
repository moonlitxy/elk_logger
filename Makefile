.PHONY: help install test run-basic run-zap docker-up docker-down clean

help: ## 显示帮助信息
	@echo "ELK Logger - Makefile命令"
	@echo ""
	@echo "可用命令："
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install: ## 安装依赖
	@echo "安装Go依赖..."
	go mod download
	go mod tidy

test: ## 运行测试
	@echo "运行测试..."
	go test -v ./...

run-basic: ## 运行基础示例
	@echo "运行基础示例..."
	cd examples/basic && go run main.go

run-zap: ## 运行Zap集成示例
	@echo "运行Zap集成示例..."
	cd examples/with_zap && go run main.go

docker-up: ## 启动ELK Docker环境
	@echo "启动Elasticsearch和Kibana..."
	docker-compose up -d
	@echo "等待服务启动..."
	@sleep 10
	@echo "Elasticsearch: http://localhost:9200"
	@echo "Kibana: http://localhost:5601"

docker-down: ## 停止ELK Docker环境
	@echo "停止ELK Docker环境..."
	docker-compose down

docker-logs: ## 查看Docker日志
	docker-compose logs -f

clean: ## 清理构建文件
	@echo "清理构建文件..."
	go clean
	rm -rf bin/
	rm -rf dist/

fmt: ## 格式化代码
	@echo "格式化代码..."
	go fmt ./...

lint: ## 运行代码检查
	@echo "运行代码检查..."
	golangci-lint run

build-examples: ## 编译所有示例
	@echo "编译示例程序..."
	go build -o bin/basic examples/basic/main.go
	go build -o bin/with_zap examples/with_zap/main.go

check-es: ## 检查Elasticsearch状态
	@echo "检查Elasticsearch状态..."
	@curl -s http://localhost:9200/_cluster/health?pretty || echo "Elasticsearch未运行"

setup-dev: docker-up install ## 设置开发环境
	@echo "开发环境设置完成！"
	@echo ""
	@echo "下一步："
	@echo "  1. 等待30秒让Elasticsearch完全启动"
	@echo "  2. 运行: make run-basic"
	@echo "  3. 在浏览器打开: http://localhost:5601"

