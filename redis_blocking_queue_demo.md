# Redis 阻塞队列命令详解

Redis 提供了多个阻塞队列命令，用于实现高效的生产者-消费者模式。

## 核心阻塞命令

### 1. BRPOP - 阻塞右侧弹出

**语法**：
```
BRPOP key [key ...] timeout
```

**功能**：
- 从列表右侧弹出元素
- 如果列表为空，阻塞等待
- 支持同时监听多个列表（按顺序检查）
- 返回第一个非空列表的元素

**示例**：
```bash
# 终端 1 - 消费者（阻塞等待）
redis-cli
> BRPOP myqueue 0
# 阻塞中...等待数据

# 终端 2 - 生产者（推送数据）
redis-cli
> LPUSH myqueue "task1"
(integer) 1

# 终端 1 立即收到：
1) "myqueue"    # 队列名
2) "task1"      # 弹出的值
```

**超时参数**：
```bash
# 0 = 永久阻塞
BRPOP myqueue 0

# 5 = 阻塞 5 秒后超时返回 nil
BRPOP myqueue 5
```

### 2. BLPOP - 阻塞左侧弹出

**语法**：
```
BLPOP key [key ...] timeout
```

**功能**：与 BRPOP 相同，但从左侧弹出

**示例**：
```bash
# FIFO 队列（先进先出）
LPUSH queue "task1"
LPUSH queue "task2"
LPUSH queue "task3"

BLPOP queue 0
# 返回: "task1" (最先进入的)
```

### 3. BRPOPLPUSH - 阻塞弹出并推入

**语法**：
```
BRPOPLPUSH source destination timeout
```

**功能**：
- 从 source 右侧弹出
- 推入 destination 左侧
- **原子操作**，保证可靠性
- Asynq 的核心命令

**示例**：
```bash
# 准备数据
LPUSH pending "task1"
LPUSH pending "task2"

# 阻塞弹出并移动
BRPOPLPUSH pending active 0
# 返回: "task1"

# 检查队列状态
LRANGE pending 0 -1
# 返回: ["task2"]

LRANGE active 0 -1
# 返回: ["task1"]
```

**可靠性保证**：
```bash
# 场景：Worker 处理任务时崩溃

# 1. 任务从 pending 移到 active
BRPOPLPUSH pending active 0
# 返回: "task1"
# pending: []
# active: ["task1"]

# 2. Worker 崩溃...

# 3. 恢复程序可以从 active 队列找到未完成的任务
LRANGE active 0 -1
# 返回: ["task1"]  # 任务还在！
```

### 4. BLMOVE - 阻塞移动（Redis 6.2+）

**语法**：
```
BLMOVE source destination LEFT|RIGHT LEFT|RIGHT timeout
```

**功能**：
- BRPOPLPUSH 的通用版本
- 可以指定源和目标的弹出/推入方向

**示例**：
```bash
# 等价于 BRPOPLPUSH
BLMOVE pending active RIGHT LEFT 5

# 从左侧弹出，推入右侧
BLMOVE pending active LEFT RIGHT 5
```

## 多队列优先级处理

### 同时监听多个队列

```bash
# 按优先级顺序监听
BRPOP high_priority default_priority low_priority 0

# 工作原理：
# 1. 先检查 high_priority
# 2. 如果为空，检查 default_priority
# 3. 如果为空，检查 low_priority
# 4. 全部为空则阻塞等待
```

**实际演示**：
```bash
# 准备数据
LPUSH low_priority "low_task"
LPUSH high_priority "high_task"

# 消费者
BRPOP high_priority default_priority low_priority 0
# 返回: 
# 1) "high_priority"  # 优先处理高优先级
# 2) "high_task"

BRPOP high_priority default_priority low_priority 0
# 返回:
# 1) "low_priority"   # 高优先级为空，处理低优先级
# 2) "low_task"
```

## 实战示例

### 示例 1：简单的任务队列

```bash
# === 生产者 ===
# 添加任务
LPUSH tasks "send_email:user1@example.com"
LPUSH tasks "process_data:file123.csv"
LPUSH tasks "generate_report:2026-02"

# === 消费者 ===
# Worker 1
BRPOP tasks 0
# 获取: "send_email:user1@example.com"

# Worker 2
BRPOP tasks 0
# 获取: "process_data:file123.csv"

# Worker 3
BRPOP tasks 0
# 获取: "generate_report:2026-02"
```

### 示例 2：可靠队列（带备份）

```bash
# === 消费者 ===
# 从 pending 获取任务，同时备份到 processing
BRPOPLPUSH pending processing 0
# 返回: "task1"
# pending: []
# processing: ["task1"]

# 处理任务...
# process_task("task1")

# 处理成功，从 processing 删除
LREM processing 1 "task1"
# processing: []

# 如果处理失败，任务仍在 processing 中
# 可以通过定时任务重新处理
```

### 示例 3：延迟队列

```bash
# 使用 Sorted Set 实现延迟队列
# score = 执行时间戳

# 添加延迟任务（5分钟后执行）
ZADD delayed_tasks 1738827600 "task1"

# 定时检查并移动到待处理队列
# (通过脚本实现)
current_time=$(date +%s)
tasks=$(redis-cli ZRANGEBYSCORE delayed_tasks 0 $current_time)
for task in $tasks; do
    redis-cli LPUSH pending "$task"
    redis-cli ZREM delayed_tasks "$task"
done

# Worker 处理
BRPOP pending 0
```

## 性能对比

### 阻塞 vs 轮询

```bash
# === 轮询方式（低效）===
while true; do
    task=$(redis-cli RPOP queue)
    if [ -z "$task" ]; then
        sleep 0.1  # CPU 空转
    else
        process_task "$task"
    fi
done

# === 阻塞方式（高效）===
while true; do
    task=$(redis-cli BRPOP queue 5)
    if [ -n "$task" ]; then
        process_task "$task"
    fi
    # 无 CPU 空转，阻塞时不消耗资源
done
```

## 常见模式

### 模式 1：FIFO 队列

```bash
# 生产者
LPUSH queue "task1"
LPUSH queue "task2"

# 消费者
BRPOP queue 0  # 先进先出
```

### 模式 2：LIFO 栈

```bash
# 生产者
LPUSH stack "task1"
LPUSH stack "task2"

# 消费者
BLPOP stack 0  # 后进先出
```

### 模式 3：工作窃取

```bash
# Worker 1 的队列
LPUSH worker1_queue "task1"
LPUSH worker1_queue "task2"

# Worker 2 可以从 Worker 1 窃取任务
BRPOPLPUSH worker1_queue worker2_queue 0
```

### 模式 4：循环队列

```bash
# 从队列获取任务，处理后放回队列尾部
BRPOPLPUSH queue queue 0
# 实现循环处理
```

## 注意事项

### 1. 超时设置

```bash
# 0 = 永久阻塞（可能导致连接长期占用）
BRPOP queue 0

# 建议使用有限超时
BRPOP queue 5  # 5秒超时，可以定期检查停止信号
```

### 2. 连接管理

```bash
# 阻塞命令会占用连接
# 确保有足够的连接池大小

# Redis 默认最大连接数
CONFIG GET maxclients
# 返回: "10000"
```

### 3. 公平性

```bash
# 多个客户端阻塞在同一队列
# Redis 按照 FIFO 顺序唤醒客户端

# Client 1: BRPOP queue 0  (先阻塞)
# Client 2: BRPOP queue 0  (后阻塞)
# 
# LPUSH queue "task1"
# -> Client 1 被唤醒
```

### 4. 原子性

```bash
# BRPOPLPUSH 是原子操作
# 不会出现任务丢失

# 非原子方式（不推荐）：
task=$(BRPOP source 0)
LPUSH destination "$task"
# 如果中间崩溃，任务丢失！

# 原子方式（推荐）：
BRPOPLPUSH source destination 0
# 一步完成，不会丢失
```

## Go 语言示例

```go
package main

import (
    "context"
    "fmt"
    "github.com/redis/go-redis/v9"
    "time"
)

func main() {
    ctx := context.Background()
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    // 生产者
    go func() {
        for i := 1; i <= 5; i++ {
            task := fmt.Sprintf("task-%d", i)
            rdb.LPush(ctx, "queue", task)
            fmt.Printf("Produced: %s\n", task)
            time.Sleep(1 * time.Second)
        }
    }()

    // 消费者
    for {
        // 阻塞等待任务，超时 5 秒
        result := rdb.BRPop(ctx, 5*time.Second, "queue")
        
        if result.Err() == redis.Nil {
            fmt.Println("Timeout, no tasks")
            continue
        }
        
        if result.Err() != nil {
            fmt.Printf("Error: %v\n", result.Err())
            break
        }
        
        // result.Val() = [queue_name, value]
        task := result.Val()[1]
        fmt.Printf("Consumed: %s\n", task)
    }
}
```

## 总结

Redis 阻塞队列命令的核心优势：

1. ✅ **高效**：无 CPU 空转，阻塞时不消耗资源
2. ✅ **即时**：任务到达立即处理，无轮询延迟
3. ✅ **可靠**：BRPOPLPUSH 原子操作保证不丢任务
4. ✅ **灵活**：支持多队列、优先级、超时控制
5. ✅ **简单**：一条命令实现复杂的队列逻辑

这就是为什么 Asynq 选择使用 Redis 阻塞队列的原因！
