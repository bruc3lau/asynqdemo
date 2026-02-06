#!/bin/bash
# Redis 阻塞队列实战演示脚本

echo "=========================================="
echo "Redis 阻塞队列命令演示"
echo "=========================================="
echo ""

# 清理之前的数据
echo "1. 清理测试数据..."
docker exec -it asynq-redis redis-cli DEL demo_queue demo_pending demo_active > /dev/null 2>&1
echo "✓ 清理完成"
echo ""

# 演示 1: BRPOP 基础用法
echo "2. 演示 BRPOP（阻塞右侧弹出）"
echo "-------------------------------------------"
echo "添加任务到队列..."
docker exec -it asynq-redis redis-cli LPUSH demo_queue "task1" "task2" "task3"
echo ""

echo "从队列弹出任务（非阻塞）："
docker exec -it asynq-redis redis-cli RPOP demo_queue
echo ""

echo "再次弹出："
docker exec -it asynq-redis redis-cli RPOP demo_queue
echo ""

echo "队列剩余："
docker exec -it asynq-redis redis-cli LRANGE demo_queue 0 -1
echo ""

# 演示 2: BRPOPLPUSH
echo "3. 演示 BRPOPLPUSH（可靠队列）"
echo "-------------------------------------------"
echo "添加任务到 pending 队列..."
docker exec -it asynq-redis redis-cli LPUSH demo_pending "email:user1" "email:user2" "email:user3"
echo ""

echo "pending 队列内容："
docker exec -it asynq-redis redis-cli LRANGE demo_pending 0 -1
echo ""

echo "使用 BRPOPLPUSH 移动任务（pending -> active）："
docker exec -it asynq-redis redis-cli BRPOPLPUSH demo_pending demo_active 0
echo ""

echo "pending 队列（任务已移除）："
docker exec -it asynq-redis redis-cli LRANGE demo_pending 0 -1
echo ""

echo "active 队列（任务已加入）："
docker exec -it asynq-redis redis-cli LRANGE demo_active 0 -1
echo ""

# 演示 3: 多队列优先级
echo "4. 演示多队列优先级处理"
echo "-------------------------------------------"
docker exec -it asynq-redis redis-cli DEL high_priority default_priority low_priority > /dev/null 2>&1

echo "添加不同优先级的任务..."
docker exec -it asynq-redis redis-cli LPUSH low_priority "low_task"
docker exec -it asynq-redis redis-cli LPUSH high_priority "high_task"
docker exec -it asynq-redis redis-cli LPUSH default_priority "default_task"
echo ""

echo "使用 BRPOP 按优先级获取任务："
echo "第一次（应该获取 high_priority）："
docker exec -it asynq-redis redis-cli BRPOP high_priority default_priority low_priority 1
echo ""

echo "第二次（应该获取 default_priority）："
docker exec -it asynq-redis redis-cli BRPOP high_priority default_priority low_priority 1
echo ""

echo "第三次（应该获取 low_priority）："
docker exec -it asynq-redis redis-cli BRPOP high_priority default_priority low_priority 1
echo ""

# 演示 4: 超时机制
echo "5. 演示超时机制"
echo "-------------------------------------------"
docker exec -it asynq-redis redis-cli DEL empty_queue > /dev/null 2>&1

echo "尝试从空队列获取任务（2秒超时）..."
echo "开始时间: $(date +%H:%M:%S)"
docker exec -it asynq-redis redis-cli BRPOP empty_queue 2
echo "结束时间: $(date +%H:%M:%S)"
echo "(应该等待约2秒后返回 nil)"
echo ""

# 查看 Asynq 实际使用的队列
echo "6. 查看 Asynq 实际队列"
echo "-------------------------------------------"
echo "Asynq 队列列表："
docker exec -it asynq-redis redis-cli KEYS "asynq:*:pending" "asynq:*:active"
echo ""

echo "查看 default 队列的待处理任务数量："
docker exec -it asynq-redis redis-cli LLEN "asynq:default:pending"
echo ""

echo "查看 default 队列的处理中任务数量："
docker exec -it asynq-redis redis-cli LLEN "asynq:default:active"
echo ""

echo "=========================================="
echo "演示完成！"
echo "=========================================="
echo ""
echo "💡 提示："
echo "1. 阻塞命令在队列为空时会等待"
echo "2. BRPOPLPUSH 是原子操作，保证任务不丢失"
echo "3. 多队列可以实现优先级处理"
echo "4. 超时参数控制最大等待时间"
echo ""
echo "🔧 进入 Redis CLI 手动测试："
echo "   docker exec -it asynq-redis redis-cli"
echo ""
