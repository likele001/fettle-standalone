#!/bin/bash
export PATH=$PATH:/usr/local/go/bin

echo "=== 编译 user-service ==="
cd /www/wwwroot/fettle/backend/user-service
go build -o /www/wwwroot/fettle/build/user-service . 2>&1
if [ $? -ne 0 ]; then
    echo "编译失败!"
    exit 1
fi
echo "编译成功"

echo "=== 停止旧进程 ==="
pkill user-service || true
sleep 1

echo "=== 加载环境变量并启动 ==="
set -a
source /www/wwwroot/fettle/.env
set +a
nohup /www/wwwroot/fettle/build/user-service > /www/wwwroot/fettle/logs/user-service.log 2>&1 &
sleep 3

echo "=== 验证进程 ==="
ps aux | grep 'build/user-service' | grep -v grep
echo ""

echo "=== 登录测试 ==="
RESP=$(curl -s -X POST 'http://localhost:9100/api/v1/admin/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13800000000","password":"Admin@2026"}')
echo "Login: $(echo $RESP | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d["code"], d["message"])' 2>/dev/null)"

TOKEN=$(echo $RESP | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["tokens"]["access_token"])' 2>/dev/null)

echo ""
echo "=== 通过 gateway 测 admin/tenants ==="
curl -s -w '\nHTTP %{http_code}\n' 'http://localhost:9100/api/v1/admin/tenants?page_size=1' \
  -H "Authorization: Bearer $TOKEN" | head -3

echo ""
echo "=== 通过 gateway 测 admin/revenue/stats ==="
curl -s -w '\nHTTP %{http_code}\n' 'http://localhost:9100/api/v1/admin/revenue/stats' \
  -H "Authorization: Bearer $TOKEN" | head -3

echo ""
echo "=== 通过 gateway 测 admin/monitor/stats ==="
curl -s -w '\nHTTP %{http_code}\n' 'http://localhost:9100/api/v1/admin/monitor/stats' \
  -H "Authorization: Bearer $TOKEN" | head -3

echo ""
echo "=== 日志 ==="
tail -5 /www/wwwroot/fettle/logs/user-service.log
