# Asynq CLI 使用指南

Asynq CLI 是一个强大的命令行工具，用于监控和管理 Asynq 任务队列。

## 安装

```bash
go install github.com/hibiken/asynq/tools/asynq@latest
```

## 常用命令

### 1. 查看统计信息

```bash
asynq stats
```

显示内容：
- **任务状态统计**：active（处理中）、pending（待处理）、scheduled（已调度）、retry（重试中）、archived（归档）、completed（已完成）
- **队列任务数量**：各个队列的任务分布
- **每日统计**：处理数量、失败数量、错误率
- **Redis 信息**：版本、运行时间、连接数、内存使用

**示例输出**：
```
Task Count by State
active  pending  scheduled  retry  archived  completed
------  -------  ---------  -----  --------  ---------
0       0        0          20     4         0

Task Count by Queue
pipeline  default
--------  -------
24        0

Daily Stats 2026-02-06 UTC
processed  failed  error rate
---------  ------  ----------
0          0       N/A

Redis Info
version  uptime  connections  memory usage
-------  ------  -----------  ------------
8.4.0    0 days  5            1.70MB
```

### 2. 队列管理

#### 列出所有队列
```bash
asynq queue ls
```

#### 查看队列详情
```bash
asynq queue inspect default
```

#### 暂停队列
```bash
asynq queue pause default
```

#### 恢复队列
```bash
asynq queue unpause default
```

#### 删除队列
```bash
asynq queue rm myqueue
```

### 3. 任务管理

#### 列出任务
```bash
# 列出待处理任务
asynq task ls --queue=default --state=pending

# 列出重试任务
asynq task ls --queue=default --state=retry

# 列出归档任务（失败任务）
asynq task ls --queue=default --state=archived

# 列出已调度任务
asynq task ls --queue=default --state=scheduled
```

#### 查看任务详情
```bash
asynq task inspect <task_id>
```

#### 取消任务
```bash
asynq task cancel <task_id>
```

#### 删除任务
```bash
asynq task delete <task_id>
```

#### 运行归档任务
```bash
# 运行所有归档任务
asynq task archive run --queue=default

# 运行指定归档任务
asynq task archive run <task_id>
```

#### 删除归档任务
```bash
# 删除所有归档任务
asynq task archive delete --queue=default

# 删除指定归档任务
asynq task archive delete <task_id>
```

### 4. 服务器管理

#### 列出所有服务器
```bash
asynq server ls
```

显示所有连接到 Redis 的 Asynq 服务器实例。

### 5. 定时任务管理

#### 列出定时任务
```bash
asynq cron ls
```

#### 查看定时任务详情
```bash
asynq cron inspect <entry_id>
```

### 6. 仪表板

#### 启动交互式仪表板
```bash
asynq dash
```

这会启动一个基于终端的交互式仪表板，实时显示：
- 队列状态
- 任务统计
- 服务器信息
- 实时更新

使用方向键导航，按 `q` 退出。

## 连接配置

### 默认连接
默认连接到 `localhost:6379`，数据库 0。

### 自定义连接

#### 使用 URI
```bash
asynq stats --uri=redis://localhost:6379/0
```

#### 使用密码
```bash
asynq stats --password=mypassword
```

#### 使用不同数据库
```bash
asynq stats --db=1
```

#### Redis Cluster
```bash
asynq stats --cluster --cluster_addrs=localhost:7000,localhost:7001,localhost:7002
```

#### TLS 连接
```bash
asynq stats --tls --tls_server=redis.example.com
```

### 配置文件

创建 `~/.asynq.yaml` 配置文件：

```yaml
uri: redis://localhost:6379/0
password: mypassword
db: 0
```

然后直接运行命令：
```bash
asynq stats
```

## 实用场景

### 场景 1：监控任务队列健康状况
```bash
# 查看整体统计
asynq stats

# 查看特定队列
asynq queue inspect default

# 实时监控
asynq dash
```

### 场景 2：处理失败任务
```bash
# 查看失败任务
asynq task ls --queue=default --state=archived

# 查看任务详情
asynq task inspect <task_id>

# 重新运行失败任务
asynq task archive run <task_id>

# 或批量重试所有失败任务
asynq task archive run --queue=default
```

### 场景 3：清理旧任务
```bash
# 删除所有归档任务
asynq task archive delete --queue=default

# 删除特定任务
asynq task delete <task_id>
```

### 场景 4：调试任务调度
```bash
# 查看已调度的任务
asynq task ls --queue=default --state=scheduled

# 查看任务详情和调度时间
asynq task inspect <task_id>
```

### 场景 5：队列维护
```bash
# 暂停队列（停止处理新任务）
asynq queue pause default

# 执行维护操作...

# 恢复队列
asynq queue unpause default
```

## 与项目集成

在我们的 `asynqdemo` 项目中使用：

```bash
# 确保 Redis 正在运行
docker-compose up -d

# 启动应用
go run main.go

# 在另一个终端查看统计
asynq stats

# 提交一些任务
python3 test_api.py

# 再次查看统计，观察变化
asynq stats

# 启动仪表板实时监控
asynq dash
```

## 常见问题

### 连接失败
```bash
Error: redis: connection refused
```
**解决方案**：确保 Redis 正在运行
```bash
docker-compose ps
docker-compose up -d
```

### 找不到队列
```bash
Error: queue not found
```
**解决方案**：队列只有在有任务时才会创建，先提交一些任务。

### 权限问题
```bash
Error: NOAUTH Authentication required
```
**解决方案**：使用 `--password` 参数或配置文件设置密码。

## 高级用法

### 批量操作
```bash
# 批量删除重试任务
asynq task ls --queue=default --state=retry | \
  awk '{print $1}' | \
  xargs -I {} asynq task delete {}
```

### 监控脚本
```bash
#!/bin/bash
# monitor.sh - 定期检查队列状态

while true; do
  echo "=== $(date) ==="
  asynq stats
  echo ""
  sleep 10
done
```

### 导出任务信息
```bash
# 导出所有待处理任务
asynq task ls --queue=default --state=pending > pending_tasks.txt
```

## 总结

Asynq CLI 提供了完整的任务队列管理功能：
- ✅ 实时监控队列状态
- ✅ 管理任务生命周期
- ✅ 处理失败任务
- ✅ 队列维护操作
- ✅ 交互式仪表板

这些工具对于生产环境的运维和调试非常有用。
