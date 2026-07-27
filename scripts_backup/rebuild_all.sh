#!/bin/bash
export PATH=$PATH:/usr/local/go/bin

echo "=== 1. 编译 user-service ==="
cd /www/wwwroot/fettle/backend/user-service
go build -o /www/wwwroot/fettle/build/user-service . 2>&1
if [ $? -ne 0 ]; then echo "user-service 编译失败"; exit 1; fi
echo "user-service 编译成功"

echo ""
echo "=== 2. 重启 user-service ==="
pkill user-service || true
sleep 1
set -a
source /www/wwwroot/fettle/.env
set +a
nohup /www/wwwroot/fettle/build/user-service > /www/wwwroot/fettle/logs/user-service.log 2>&1 &
sleep 3
echo "user-service 进程:"
ps aux | grep 'build/user-service' | grep -v grep

echo ""
echo "=== 3. 编译前端 ==="
cd /www/wwwroot/fettle/frontend/web-platform
npm run build 2>&1 | tail -5

echo ""
echo "=== 4. 测试活跃租户接口 ==="
TOKEN=$(curl -s -X POST 'http://localhost:9100/api/v1/admin/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13800000000","password":"Admin@2026"}' | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["tokens"]["access_token"])' 2>/dev/null)

echo "--- active-tenants ---"
curl -s "http://localhost:9100/api/v1/admin/revenue/active-tenants" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo ""
echo "--- revenue/trend ---"
curl -s "http://localhost:9100/api/v1/admin/revenue/trend" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo ""
echo "=== 5. 日志 ==="
tail -3 /www/wwwroot/fettle/logs/user-service.log
