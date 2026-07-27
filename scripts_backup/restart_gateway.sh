#!/bin/bash
export PATH=$PATH:/usr/local/go/bin

echo "=== 编译 gateway ==="
cd /www/wwwroot/fettle/backend/gateway
go build -o /www/wwwroot/fettle/build/gateway . 2>&1
if [ $? -ne 0 ]; then
    echo "编译失败!"
    exit 1
fi
echo "编译成功"

echo "=== 停止旧进程 ==="
pkill gateway || true
sleep 1

echo "=== 加载环境变量并启动 ==="
set -a
source /www/wwwroot/fettle/.env
set +a
nohup /www/wwwroot/fettle/build/gateway > /www/wwwroot/fettle/logs/gateway.log 2>&1 &
sleep 3

echo "=== 验证进程 ==="
ps aux | grep 'build/gateway' | grep -v grep
echo ""

echo "=== 测试无尾斜杠 /admin/tenants ==="
curl -s -o /dev/null -w 'HTTP %{http_code} (redirect: %{redirect_url})\n' 'http://localhost:9100/api/v1/admin/tenants?page_size=1' -H 'Authorization: Bearer test'

echo "=== 测试有尾斜杠 /admin/tenants/ ==="
curl -s -o /dev/null -w 'HTTP %{http_code} (redirect: %{redirect_url})\n' 'http://localhost:9100/api/v1/admin/tenants/?page_size=1' -H 'Authorization: Bearer test'

echo "=== 测试 /admin/revenue/stats ==="
curl -s -o /dev/null -w 'HTTP %{http_code} (redirect: %{redirect_url})\n' 'http://localhost:9100/api/v1/admin/revenue/stats' -H 'Authorization: Bearer test'

echo "=== 测试 /admin/monitor/health ==="
curl -s -o /dev/null -w 'HTTP %{http_code} (redirect: %{redirect_url})\n' 'http://localhost:9100/api/v1/admin/monitor/health' -H 'Authorization: Bearer test'

echo "=== 测试公开接口 /admin/auth/login ==="
curl -s -o /dev/null -w 'HTTP %{http_code} (redirect: %{redirect_url})\n' -X POST 'http://localhost:9100/api/v1/admin/auth/login' -H 'Content-Type: application/json' -d '{"phone":"test","password":"test"}'

echo ""
echo "=== gateway 日志 ==="
tail -5 /www/wwwroot/fettle/logs/gateway.log
