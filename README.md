# Asynq Demo - Gin + Asynq 任务调度示例

这是一个使用 Go 语言实现的简单任务调度系统，展示了如何使用 Gin 框架和 Asynq 库来构建异步任务处理系统。

## 功能特性

- ✅ 使用 **Gin** 框架提供 RESTful API
- ✅ 使用 **Asynq** 进行异步任务处理
- ✅ 支持多种任务类型（邮件发送、数据处理）
- ✅ 任务延迟执行支持
- ✅ 并发任务处理
- ✅ Python 测试脚本

## 项目结构

```
asynqdemo/
├── main.go              # 主程序入口
├── server/              # HTTP 服务器
│   └── server.go        # Gin 路由和处理器
├── worker/              # Asynq 工作器
│   └── worker.go        # 任务处理器注册
├── tasks/               # 任务定义
│   └── tasks.go         # 任务类型和处理函数
├── docker-compose.yml   # Redis 容器配置
├── test_api.py          # Python 测试脚本
└── README.md            # 本文件
```

## 前置要求

- Go 1.21 或更高版本
- Redis 服务器
- Python 3.7+ (用于测试脚本)
- Docker 和 Docker Compose (可选，用于运行 Redis)

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 启动 Redis

使用 Docker Compose:

```bash
docker-compose up -d
```

或者使用本地 Redis:

```bash
redis-server
```

### 3. 运行应用

```bash
go run main.go
```

应用将启动两个服务：
- HTTP 服务器: `http://localhost:8081`
- Asynq 工作器: 连接到 Redis 并处理任务

### 4. 测试 API

使用 Python 测试脚本:

```bash
python3 test_api.py
```

或使用 curl:

```bash
# 健康检查
curl http://localhost:8080/api/health

# 提交邮件任务
curl -X POST http://localhost:8080/api/tasks/email \
  -H "Content-Type: application/json" \
  -d '{
    "to": "user@example.com",
    "subject": "测试邮件",
    "body": "这是一封测试邮件"
  }'

# 提交数据处理任务
curl -X POST http://localhost:8080/api/tasks/process \
  -H "Content-Type: application/json" \
  -d '{
    "data_id": "DATA-001",
    "action": "transform",
    "delay": 3
  }'
```

## API 端点

### 健康检查

```
GET /api/health
```

响应示例:
```json
{
  "status": "healthy",
  "time": "2026-02-05T18:17:00+08:00"
}
```

### 提交邮件任务

```
POST /api/tasks/email
```

请求体:
```json
{
  "to": "user@example.com",
  "subject": "邮件主题",
  "body": "邮件内容"
}
```

响应示例:
```json
{
  "task_id": "abc123...",
  "queue": "default",
  "message": "Email task submitted successfully"
}
```

### 提交数据处理任务

```
POST /api/tasks/process
```

请求体:
```json
{
  "data_id": "DATA-001",
  "action": "transform",
  "delay": 3
}
```

响应示例:
```json
{
  "task_id": "def456...",
  "queue": "default",
  "message": "Data processing task submitted successfully",
  "process_at": "in 5 seconds"
}
```

## 架构说明

### 组件

1. **Gin HTTP 服务器** (`server/server.go`)
   - 提供 RESTful API 接口
   - 接收任务提交请求
   - 将任务加入 Asynq 队列

2. **Asynq 工作器** (`worker/worker.go`)
   - 从 Redis 队列中获取任务
   - 并发执行任务
   - 支持任务重试机制

3. **任务定义** (`tasks/tasks.go`)
   - 定义任务类型和数据结构
   - 实现任务处理逻辑

### 工作流程

```
客户端请求 → Gin API → Asynq Client → Redis 队列
                                          ↓
                                    Asynq Worker → 执行任务
```

1. 客户端通过 HTTP API 提交任务
2. Gin 服务器将任务序列化并加入 Redis 队列
3. Asynq 工作器从队列中获取任务
4. 工作器执行任务处理逻辑
5. 任务完成或失败（支持重试）

## 配置

主要配置项在 `main.go` 中:

```go
const (
    redisAddr  = "localhost:6379"  // Redis 地址
    serverAddr = ":8080"            // HTTP 服务器端口
)
```

工作器配置在 `worker/worker.go` 中:

```go
asynq.Config{
    Concurrency: 10,  // 并发工作器数量
}
```

## 开发说明

### 添加新任务类型

1. 在 `tasks/tasks.go` 中定义任务类型常量
2. 创建任务数据结构
3. 实现任务创建函数
4. 实现任务处理函数
5. 在 `worker/worker.go` 中注册处理器
6. 在 `server/server.go` 中添加 API 端点

### 示例：添加短信任务

```go
// tasks/tasks.go
const TypeSMSDelivery = "sms:delivery"

type SMSPayload struct {
    Phone   string `json:"phone"`
    Message string `json:"message"`
}

func HandleSMSDeliveryTask(ctx context.Context, t *asynq.Task) error {
    var p SMSPayload
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        return err
    }
    // 实现短信发送逻辑
    return nil
}
```

## 监控和调试

查看应用日志以监控任务执行:

```bash
go run main.go
```

日志示例:
```
🚀 Starting Asynq Demo Application
========================================
🚀 Starting Asynq worker server...
📡 Connected to Redis at: localhost:6379
👷 Worker is ready to process tasks
🌐 Starting HTTP server on :8080
📨 Enqueued email task: ID=abc123 Queue=default
📧 [Email Task] Sending email to: user@example.com
✅ [Email Task] Successfully sent email to: user@example.com
```

## 故障排除

### Redis 连接失败

确保 Redis 正在运行:
```bash
docker-compose ps
# 或
redis-cli ping
```

### 端口已被占用

修改 `main.go` 中的 `serverAddr` 常量:
```go
const serverAddr = ":8081"  // 使用其他端口
```

### 任务未执行

1. 检查工作器是否正常运行
2. 查看日志中是否有错误信息
3. 确认 Redis 连接正常

## 生产环境建议

- 使用环境变量配置 Redis 地址和服务器端口
- 添加认证和授权机制
- 实现任务状态查询 API
- 配置 Asynq 的重试策略和超时设置
- 使用 Redis Sentinel 或 Cluster 提高可用性
- 添加监控和告警（Prometheus + Grafana）
- 实现优雅关闭机制

## 许可证

MIT License
