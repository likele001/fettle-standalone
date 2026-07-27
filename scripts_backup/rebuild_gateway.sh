#!/bin/bash
export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin

echo "=== 编译 gateway ==="
cd /www/wwwroot/fettle/backend/gateway
go build -o /www/wwwroot/fettle/build/gateway . 2>&1
if [ $? -ne 0 ]; then echo "编译失败"; exit 1; fi
echo "编译成功"

echo "=== 重启 gateway ==="
kill $(ps aux | grep '[g]ateway' | grep -v node | awk '{print $2}') 2>/dev/null
sleep 1
cd /www/wwwroot/fettle
set -a
source .env
set +a
nohup ./build/gateway > logs/gateway.log 2>&1 &
sleep 2

echo "=== 检查进程 ==="
ps aux | grep '[g]ateway' | grep -v node

echo "=== 健康检查 ==="
curl -s http://localhost:9100/health
echo ""

echo "=== 测试 admin billing ==="
TOKEN=$(curl -s -X POST 'http://localhost:9100/api/v1/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13800000000","password":"Admin@2026"}' | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["tokens"]["access_token"])' 2>/dev/null)

echo "GET /billing/plans:"
curl -s "http://localhost:9100/api/v1/billing/plans" -H "Authorization: Bearer $TOKEN" | python3 -c 'import sys,json; d=json.load(sys.stdin); print("OK" if d.get("code")==0 else d)' 2>/dev/null

echo "GET /billing/subscriptions:"
curl -s "http://localhost:9100/api/v1/billing/subscriptions" -H "Authorization: Bearer $TOKEN" | python3 -c 'import sys,json; d=json.load(sys.stdin); print("OK" if d.get("code")==0 or isinstance(d, dict) else d)' 2>/dev/null

echo "GET /billing/quota:"
curl -s "http://localhost:9100/api/v1/billing/quota" -H "Authorization: Bearer $TOKEN" | python3 -c 'import sys,json; d=json.load(sys.stdin); print("OK" if d.get("code")==0 or isinstance(d, dict) else d)' 2>/dev/null

echo "GET /billing/records:"
curl -s "http://localhost:9100/api/v1/billing/records?page=1&page_size=10" -H "Authorization: Bearer $TOKEN" | python3 -c 'import sys,json; d=json.load(sys.stdin); print("OK" if d.get("code")==0 or isinstance(d, dict) else d)' 2>/dev/null

echo "=== 完成 ==="
