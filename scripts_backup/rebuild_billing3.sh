#!/bin/bash
export PATH=$PATH:/usr/local/go/bin

echo "=== 上传并编译 billing-service ==="
cd /www/wwwroot/fettle/backend/billing-service
go build -o /www/wwwroot/fettle/build/billing-service . 2>&1
if [ $? -ne 0 ]; then echo "编译失败"; exit 1; fi
echo "编译成功"

echo ""
echo "=== 重启 billing-service ==="
pkill billing-service || true
sleep 1
set -a
source /www/wwwroot/fettle/.env
set +a
nohup /www/wwwroot/fettle/build/billing-service > /www/wwwroot/fettle/logs/billing-service.log 2>&1 &
sleep 3

echo "=== 进程 ==="
ps aux | grep billing-service | grep -v grep

echo ""
echo "=== 日志 ==="
tail -5 /www/wwwroot/fettle/logs/billing-service.log

echo ""
echo "=== 测试 API ==="
TOKEN=$(curl -s -X POST 'http://localhost:9100/api/v1/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13800000000","password":"Admin@2026"}' | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["tokens"]["access_token"])' 2>/dev/null)

echo "--- 直接测 billing-service /billing/plans ---"
curl -s http://localhost:9600/billing/plans -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo ""
echo "--- 通过 gateway /api/v1/billing/plans ---"
curl -s 'http://localhost:9100/api/v1/billing/plans' \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo ""
echo "--- /billing/subscriptions ---"
curl -s 'http://localhost:9100/api/v1/billing/subscriptions' \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo ""
echo "--- /billing/quota ---"
curl -s 'http://localhost:9100/api/v1/billing/quota' \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo ""
echo "--- /billing/records ---"
curl -s 'http://localhost:9100/api/v1/billing/records' \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null
