#!/usr/bin/env python3
"""
Asynq Demo API 测试脚本
用于测试 Gin + Asynq 任务调度系统的 HTTP API
"""

import requests
import json
import time
from typing import Dict, Any

# API 基础 URL
BASE_URL = "http://localhost:3000/api"

def print_section(title: str):
    """打印分隔线"""
    print("\n" + "=" * 60)
    print(f"  {title}")
    print("=" * 60)

def check_health() -> bool:
    """检查服务健康状态"""
    try:
        response = requests.get(f"{BASE_URL}/health", timeout=5)
        if response.status_code == 200:
            data = response.json()
            print(f"✅ 服务健康检查通过")
            print(f"   状态: {data.get('status')}")
            print(f"   时间: {data.get('time')}")
            return True
        else:
            print(f"❌ 健康检查失败: HTTP {response.status_code}")
            return False
    except requests.exceptions.RequestException as e:
        print(f"❌ 无法连接到服务: {e}")
        return False

def submit_email_task(to: str, subject: str, body: str) -> Dict[str, Any]:
    """提交邮件任务"""
    payload = {
        "to": to,
        "subject": subject,
        "body": body
    }
    
    try:
        response = requests.post(
            f"{BASE_URL}/tasks/email",
            json=payload,
            headers={"Content-Type": "application/json"},
            timeout=5
        )
        
        if response.status_code == 200:
            data = response.json()
            print(f"✅ 邮件任务提交成功")
            print(f"   任务ID: {data.get('task_id')}")
            print(f"   队列: {data.get('queue')}")
            print(f"   收件人: {to}")
            return data
        else:
            print(f"❌ 任务提交失败: HTTP {response.status_code}")
            print(f"   响应: {response.text}")
            return {}
    except requests.exceptions.RequestException as e:
        print(f"❌ 请求失败: {e}")
        return {}

def submit_data_process_task(data_id: str, action: str, delay: int = 2) -> Dict[str, Any]:
    """提交数据处理任务"""
    payload = {
        "data_id": data_id,
        "action": action,
        "delay": delay
    }
    
    try:
        response = requests.post(
            f"{BASE_URL}/tasks/process",
            json=payload,
            headers={"Content-Type": "application/json"},
            timeout=5
        )
        
        if response.status_code == 200:
            data = response.json()
            print(f"✅ 数据处理任务提交成功")
            print(f"   任务ID: {data.get('task_id')}")
            print(f"   队列: {data.get('queue')}")
            print(f"   数据ID: {data_id}")
            print(f"   操作: {action}")
            print(f"   延迟: {delay}秒")
            return data
        else:
            print(f"❌ 任务提交失败: HTTP {response.status_code}")
            print(f"   响应: {response.text}")
            return {}
    except requests.exceptions.RequestException as e:
        print(f"❌ 请求失败: {e}")
        return {}

def main():
    """主测试函数"""
    print_section("Asynq Demo API 测试")
    
    # 1. 健康检查
    print_section("1. 健康检查")
    if not check_health():
        print("\n⚠️  服务未运行，请先启动服务:")
        print("   1. 启动 Redis: docker-compose up -d")
        print("   2. 启动应用: go run main.go")
        return
    
    # 2. 提交邮件任务
    print_section("2. 提交邮件任务")
    submit_email_task(
        to="user@example.com",
        subject="测试邮件",
        body="这是一封测试邮件，用于验证 Asynq 任务调度系统"
    )
    
    time.sleep(1)
    
    submit_email_task(
        to="admin@example.com",
        subject="系统通知",
        body="您的账户已成功激活"
    )
    
    # 3. 提交数据处理任务
    print_section("3. 提交数据处理任务")
    submit_data_process_task(
        data_id="DATA-001",
        action="transform",
        delay=3
    )
    
    time.sleep(1)
    
    submit_data_process_task(
        data_id="DATA-002",
        action="analyze",
        delay=5
    )
    
    # 4. 批量提交任务
    print_section("4. 批量提交任务")
    print("提交 5 个邮件任务...")
    for i in range(5):
        submit_email_task(
            to=f"user{i+1}@example.com",
            subject=f"批量邮件 #{i+1}",
            body=f"这是第 {i+1} 封批量邮件"
        )
        time.sleep(0.3)
    
    print_section("测试完成")
    print("\n💡 提示:")
    print("   - 查看服务端日志以观察任务执行情况")
    print("   - 任务会在后台异步执行")
    print("   - 数据处理任务会在提交后 5 秒开始执行")
    print()

if __name__ == "__main__":
    main()
